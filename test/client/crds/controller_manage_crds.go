package main

import (
	"context"
	"github.com/go-logr/logr"
	devopsV1 "github.com/tomoncle/k8s-operator-nginx/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that `exec-entrypoint` and `run` can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	_scheme  = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	//  registered for the type v1.Nginx in scheme
	utilruntime.Must(devopsV1.AddToScheme(_scheme))
}

type FakeReconcile struct {
	client.Client
	scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *FakeReconcile) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithName("Reconcile")

	var nginx devopsV1.Nginx
	if err := r.Get(ctx, req.NamespacedName, &nginx); err != nil {
		logger.Error(err, "查询Nginx失败！")
		return ctrl.Result{}, err
	}
	logger.Info("查询结果：", "nginx", nginx.Name, "GVK", nginx.GroupVersionKind())
	return ctrl.Result{}, nil
}

func main() {
	ctrl.SetLogger(zap.New())
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             _scheme,
		MetricsBindAddress: ":8070",
	})
	if err != nil {
		setupLog.Error(err, "创建Manager失败！")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&devopsV1.Nginx{}).
		Complete(&FakeReconcile{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			Log:    ctrl.Log.WithName("test").WithName("test_crd"),
		})
	if err != nil {
		setupLog.Error(err, "创建ControllerManage失败")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "启动manager失败")
		os.Exit(1)
	}
}
