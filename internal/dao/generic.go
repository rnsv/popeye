package dao

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Generic represents a generic resource.
type Generic struct {
	NonResource
}

// List returns a collection of resources.
func (g *Generic) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labelSel, ok := ctx.Value(internal.KeyLabels).(string)
	if !ok {
		log.Debug().Msgf("No label selector found in context. Listing all resources")
	}
	if client.IsAllNamespace(ns) {
		ns = client.AllNamespaces
	}

	var (
		ll  *unstructured.UnstructuredList
		err error
	)
	if client.IsClusterScoped(ns) {
		ll, err = g.dynClient().List(ctx, metav1.ListOptions{LabelSelector: labelSel})
	} else {
		ll, err = g.dynClient().Namespace(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSel})
	}
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ll.Items))
	for i := range ll.Items {
		oo[i] = &ll.Items[i]
	}

	return oo, nil
}

// Get returns a given resource.
func (g *Generic) Get(ctx context.Context, path string) (runtime.Object, error) {
	var opts metav1.GetOptions

	ns, n := client.Namespaced(path)
	dial := g.dynClient()
	if client.IsClusterScoped(ns) {
		return dial.Get(ctx, n, opts)
	}

	return dial.Namespace(ns).Get(ctx, n, opts)
}

func (g *Generic) dynClient() dynamic.NamespaceableResourceInterface {
	return g.Client().DynDialOrDie().Resource(g.gvr.GVR())
}
