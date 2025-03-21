/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	elasticwebv1 "elasticweb/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// deployment中的APP标签
	APP_NAME = "elastic-app"
	// 单个POD的CPU资源申请
	CPU_REQUEST = "100m"
	// 单个POD的CPU资源上限
	CPU_LIMIT = "100m"
	// 单个POD的内存资源申请
	MEM_REQUEST = "512Mi"
	// 单个POD的内存资源上限
	MEM_LIMIT = "512Mi"
)

var (
	log = ctrl.Log.WithName("setup")
)

// ElasticWebReconciler reconciles a ElasticWeb object
type ElasticWebReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=elasticweb.com.bolingcavalry,resources=elasticwebs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticweb.com.bolingcavalry,resources=elasticwebs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elasticweb.com.bolingcavalry,resources=elasticwebs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking,resources=ingresss,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ElasticWeb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *ElasticWebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// your logic here
	log.Info("1. start reconcile logic")

	//实例化数据结构
	instance := &elasticwebv1.ElasticWeb{}

	// 通过客户端工具查询，查询条件是
	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {
		//如果没有实例，就返回空，这样外部就不再立即调用Reconcile方法了
		if errors.IsNotFound(err) {
			log.Info("2.1 instance not found, maybe removed")
			return reconcile.Result{}, nil
		}

		log.Error(err, "2.2 error")
		return ctrl.Result{}, err
	}

	log.Info("3. instance: " + instance.String())

	// 查找deployment
	deployment := &appsv1.Deployment{}

	// 用客户端工具查询
	err = r.Get(ctx, req.NamespacedName, deployment)

	// 查找时发生异常，以及查出来没有结果的处理逻辑
	if err != nil {
		// 如果没有实例就要创建了
		if errors.IsNotFound(err) {
			log.Info("4. deployment not exists")

			// 如果对QPS没有需求，此时又没有deployment，就啥事都不做了
			if *(instance.Spec.TotalQPS) < 1 {
				log.Info("5.1 not need deployment")
				return ctrl.Result{}, nil
			}

			// 先要创建service
			if err = createServiceIfNotExists(ctx, r, instance, req); err != nil {
				log.Error(err, "5.2 error")
				return ctrl.Result{}, err
			}

			// 立即创建deployment
			if err = createDeployment(ctx, r, instance); err != nil {
				log.Error(err, "5.3 error")
				return ctrl.Result{}, err
			}

			// 如果创建成功就更新状态
			if err = updateStatus(ctx, r, instance); err != nil {
				log.Error(err, "5.4 update status error")
				return ctrl.Result{}, nil
			}

			// 创建成功就可以返回了
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "7. error")
			return ctrl.Result{}, err
		}
	}

	// 如果查到了deployment，并且没有返回错误，就走下面的逻辑
	// 根据单QPS和总QPS计算期望的副本数
	expectReplicas := getExpectReplicas(instance)

	// 当前deployment的期望副本数
	realReplicas := *deployment.Spec.Replicas

	log.Info(fmt.Sprintf("9. expectReplicas [%d], realReplicas [%d]", expectReplicas, realReplicas))
	// log.Info("如果expectReplicas和realReplicas相等，就直接返回了")
	// // 如果expectReplicas和realReplicas相等，就直接返回了
	// if expectReplicas == realReplicas {
	// 	log.Info("10. return now")
	// 	return ctrl.Result{}, nil
	// }
	if expectReplicas != realReplicas {

		log.Info("-----------------11----------------------")
		// 如果expectReplicas和realReplicas不相等，就需要调整。
		*(deployment.Spec.Replicas) = expectReplicas
		log.Info("11. update deployment`s Replicas")

		// 通过客户端更新deployment
		if err = r.Update(ctx, deployment); err != nil {
			log.Error(err, "12. update deployment replicas error")
			return ctrl.Result{}, err
		}

		log.Info("13. update status")

		// 如果更新deployment的Replicas成功，就更新状态
		if err = updateStatus(ctx, r, instance); err != nil {
			log.Error(err, "14. update status error")
			return ctrl.Result{}, err
		}
	}
	var needUpdate bool
	deployment, needUpdate = getDiffDeployment(ctx, instance, deployment)

	if needUpdate {
		if err = r.Update(ctx, deployment); err != nil {
			log.Error(err, "15. update deployment replicas error")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticWebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticwebv1.ElasticWeb{}).
		Named("elasticweb").
		Complete(r)
}

func getExpectReplicas(elasticWeb *elasticwebv1.ElasticWeb) int32 {
	// 单POD的QPS
	singlePodQPS := *(elasticWeb.Spec.SinglePodQPS)

	// 期望的总QPS
	totalQPS := *(elasticWeb.Spec.TotalQPS)

	replicas := totalQPS / singlePodQPS

	if totalQPS%singlePodQPS > 0 {
		replicas++
	}
	return replicas
}

// 1.先查看service是否存在，不存在才创建；
// 2.将service和CRD实例elasticWeb建立关联(controllerutil.SetControllerReference方法)，这样当elasticWeb被删除的时候，service会被自动删除而无需我们干预；
// 3.创建service的时候用到了client-go工具，推荐您阅读《client-go实战系列》,工具越熟练，编码越尽兴；
func createServiceIfNotExists(ctx context.Context, r *ElasticWebReconciler, elasticWeb *elasticwebv1.ElasticWeb, req ctrl.Request) error {
	service := &corev1.Service{}
	err := r.Get(ctx, req.NamespacedName, service)

	// 如果查询结果没有错误，证明service正常，就不做任何操作
	if err == nil {
		log.Info("[%s] service exists", service)
	}

	// 如果错误不是NotFound，就返回错误
	if !errors.IsNotFound(err) {
		log.Error(err, "query service error")
		return err
	}

	// 实例化service ports
	var SvcPorts []corev1.ServicePort
	for _, v := range elasticWeb.Spec.Service.Ports {
		tmp := corev1.ServicePort{
			Name:       v.Name,
			Protocol:   corev1.ProtocolTCP,
			Port:       *v.Port,
			TargetPort: intstr.FromInt(int(*v.TargetPort)),
		}
		SvcPorts = append(SvcPorts, tmp)
	}

	// 实例化一个数据结构
	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: elasticWeb.Namespace,
			Name:      elasticWeb.Name,
		},
		Spec: corev1.ServiceSpec{
			Ports: SvcPorts,
			Type:  corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": elasticWeb.Name,
			},
		},
	}

	// 这一步非常关键
	// 建立关联后，删除elasticweb资源时，就会将deployment也删除掉
	log.Info("set reference")
	if err := controllerutil.SetControllerReference(elasticWeb, service, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	// 创建service
	log.Info("Start create service")
	if err := r.Create(ctx, service); err != nil {
		log.Error(err, "create service error")
		return err
	}
	log.Info("create service success")
	return nil
}

// 新建deployment
func createDeployment(ctx context.Context, r *ElasticWebReconciler, elasticWeb *elasticwebv1.ElasticWeb) error {

	// 计算期望的POD数量
	expectReplicas := getExpectReplicas(elasticWeb)

	log.Info(fmt.Sprintf("expectReplicas [%d]", expectReplicas))

	// 实例化containers

	var containers []corev1.Container
	for _, cv := range elasticWeb.Spec.Deploy {
		var tmpPorts []corev1.ContainerPort
		for _, cv1 := range cv.Ports {
			tmpPorts = append(tmpPorts, corev1.ContainerPort{
				Name:          cv1.Name,
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: *cv1.Port,
			})

		}
		tmp := corev1.Container{
			Name:            cv.Name,
			Image:           cv.Image,
			ImagePullPolicy: "IfNotPresent",
			Ports:           tmpPorts,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse(CPU_REQUEST),
					"memory": resource.MustParse(MEM_REQUEST),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse(CPU_LIMIT),
					"memory": resource.MustParse(MEM_LIMIT),
				},
			},
			// ReadinessProbe: &corev1.Probe{
			// 	InitialDelaySeconds: 10,
				
			// },
		}
		containers = append(containers, tmp)
	}

	// 实例化一个数据结构
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: elasticWeb.Namespace,
			Name:      elasticWeb.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(expectReplicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": APP_NAME,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": APP_NAME,
					},
				},
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}

	// 这一步非常关键！
	// 建立关联后，删除elasticweb资源时就会将deployment也删除掉
	log.Info("set reference")
	if err := controllerutil.SetControllerReference(elasticWeb, deployment, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	// 创建deployment
	log.Info("start create deployment")
	if err := r.Create(ctx, deployment); err != nil {
		log.Error(err, "create deployment error")
		return err
	}

	log.Info("create deployment success")
	return nil
}

// 完成pod的处理后，更新最新状态
func updateStatus(ctx context.Context, r *ElasticWebReconciler, elasticWeb *elasticwebv1.ElasticWeb) error {

	// 单个pod的QPS
	singlePodQPS := *(elasticWeb.Spec.SinglePodQPS)

	// pod总数

	replicas := getExpectReplicas(elasticWeb)

	// 当pod创建完毕后，当前系统实际的QPS：单个pod的QPS * pod总数
	// 如果该字段还没有初始化，就先做初始化
	if nil == elasticWeb.Status.RealQPS {
		elasticWeb.Status.RealQPS = new(int32)
	}
	*(elasticWeb.Status.RealQPS) = singlePodQPS * replicas
	log.Info(fmt.Sprintf("singlePodQPS [%d],replicas [%d],realQPS [%d]", singlePodQPS, replicas, *(&elasticWeb.Status.RealQPS)))

	if err := r.Update(ctx, elasticWeb); err != nil {
		log.Error(err, "update instance error")
		return err
	}

	return nil
}

func getDiffDeployment(ctx context.Context, elasticWeb *elasticwebv1.ElasticWeb, oldDeployment *appsv1.Deployment) (newDeployment *appsv1.Deployment, needUpdate bool) {
	// 当前deployment容器信息
	containers := oldDeployment.Spec.Template.Spec.Containers
	needUpdate = false
	for i1, v1 := range containers {
		for _, v2 := range elasticWeb.Spec.Deploy {
			if v1.Name == v2.Name && v1.Image != v2.Image {
				oldDeployment.Spec.Template.Spec.Containers[i1].Image = v2.Image
				log.Info("15. set deployment image")
				needUpdate = true
			}
		}
	}
	return oldDeployment, needUpdate
}
