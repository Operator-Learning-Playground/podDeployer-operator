package main

import (
	podrestarterv1alpha1 "github.com/myoperator/poddeployer/pkg/apis/podDeployer/v1alpha1"
	"github.com/myoperator/poddeployer/pkg/controller"
	"github.com/myoperator/poddeployer/pkg/k8sconfig"
	v1 "k8s.io/api/apps/v1"
	_ "k8s.io/code-generator"
	"k8s.io/klog/v2"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

/*
	manager 主要用来管理Controller Admission Webhook 包括：
	访问资源对象的client cache scheme 并提供依赖注入机制 优雅关闭机制

	operator = crd + controller + webhook
*/

func main() {

	logf.SetLogger(zap.New())
	var d time.Duration = 0
	// 1. 管理器初始化
	mgr, err := manager.New(k8sconfig.K8sRestConfig(), manager.Options{
		Logger: logf.Log.WithName("poddeployer-operator"),
		SyncPeriod: &d, // resync不设置触发
	})
	if err != nil {
		mgr.GetLogger().Error(err, "unable to set up manager")
		os.Exit(1)
	}

	// 2. ++ 注册进入序列化表
	err = podrestarterv1alpha1.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		klog.Error(err, "unable add schema")
		os.Exit(1)
	}

	// 3. 控制器相关
	podReStarterCtl := controller.NewPodDeployerController()

	err = builder.ControllerManagedBy(mgr).
		For(&podrestarterv1alpha1.Poddeployer{}).
		Watches(&source.Kind{Type: &v1.Deployment{}},
			handler.Funcs{
				DeleteFunc: podReStarterCtl.DeploymentDeleteHandler,
			},
		).
		Complete(podReStarterCtl)

	errC := make(chan error)

	// 4. 启动controller管理器
	go func() {
		klog.Info("controller start!! ")
		if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
			errC <- err
		}
	}()

	// 这里会阻塞
	getError := <-errC
	log.Println(getError.Error())

}
