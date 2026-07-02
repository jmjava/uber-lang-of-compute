package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

const (
	DefaultTSDBPort     = 9090
	DefaultTSDBEndpoint = "http://127.0.0.1:9090"
)

// ResolveConfig determines store configuration from workflow provisioning and optional ComputeContext.
type ResolveConfig struct {
	StorePath          string
	StoreType          Type
	StoreEndpoint      string
	ComputeContextRef  string
	Namespace          string
	StoreRoot          string
}

// OpenResolved opens the appropriate backend for a workflow execution.
func OpenResolved(ctx context.Context, c client.Client, cfg ResolveConfig) (Backend, error) {
	storeType, endpoint, path := cfg.StoreType, cfg.StoreEndpoint, cfg.StorePath

	if cfg.ComputeContextRef != "" && c != nil {
		var computeCtx kblv1alpha1.ComputeContext
		if err := c.Get(ctx, client.ObjectKey{Name: cfg.ComputeContextRef}, &computeCtx); err == nil {
			if computeCtx.Spec.StoreType == string(TypeTSDB) {
				storeType = TypeTSDB
				if computeCtx.Spec.StoreEndpoint != "" {
					endpoint = computeCtx.Spec.StoreEndpoint
				} else if computeCtx.Status.StoreEndpoint != "" {
					endpoint = computeCtx.Status.StoreEndpoint
				} else {
					endpoint = DefaultTSDBEndpoint
				}
			} else if computeCtx.Spec.StorePath != "" {
				path = computeCtx.Spec.StorePath
			}
		}
	}

	if storeType == TypeTSDB {
		if endpoint == "" {
			endpoint = DefaultTSDBEndpoint
		}
		return OpenBackend(Config{Type: TypeTSDB, Endpoint: endpoint})
	}

	if path == "" {
		root := cfg.StoreRoot
		if root == "" {
			root = "/var/kbl/store"
		}
		path = filepath.Join(root, cfg.Namespace, "default.db")
	}
	return OpenBackend(Config{Type: TypeSQLite, Path: path})
}

// ConfigFromWorkflow builds resolve config from a Workflow CR.
func ConfigFromWorkflow(wf *kblv1alpha1.Workflow, storeRoot string) ResolveConfig {
	path := wf.Spec.Provisioning.StorePath
	storeType := TypeSQLite
	endpoint := ""

	if wf.Spec.Provisioning.StorePath != "" && isTSDBPath(wf.Spec.Provisioning.StorePath) {
		storeType = TypeTSDB
		endpoint = wf.Spec.Provisioning.StorePath
		path = ""
	}

	return ResolveConfig{
		StorePath:         path,
		StoreType:         storeType,
		StoreEndpoint:     endpoint,
		ComputeContextRef: wf.Spec.Routing.ComputeContextRef,
		Namespace:         wf.Namespace,
		StoreRoot:         storeRoot,
	}
}

func isTSDBPath(p string) bool {
	return strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://")
}

// ConfigFromDominoChain builds resolve config from a DominoChain CR.
func ConfigFromDominoChain(chain *kblv1alpha1.DominoChain, storeRoot string) ResolveConfig {
	storeType := TypeSQLite
	endpoint := ""
	path := chain.Spec.StorePath

	if isTSDBPath(path) {
		storeType = TypeTSDB
		endpoint = path
		path = ""
	}

	return ResolveConfig{
		StorePath:     path,
		StoreType:     storeType,
		StoreEndpoint: endpoint,
		Namespace:     chain.Namespace,
		StoreRoot:     storeRoot,
	}
}

func OpenForDominoChain(ctx context.Context, c client.Client, chain *kblv1alpha1.DominoChain, storeRoot string) (Backend, error) {
	cfg := ConfigFromDominoChain(chain, storeRoot)
	if cfg.StorePath == "" && cfg.StoreType != TypeTSDB {
		cfg.StorePath = filepath.Join(storeRoot, chain.Namespace, chain.Name+".db")
	}
	b, err := OpenResolved(ctx, c, cfg)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return b, nil
}
// OpenForWorkflow is a convenience wrapper.
func OpenForWorkflow(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow, storeRoot string) (Backend, error) {
	cfg := ConfigFromWorkflow(wf, storeRoot)
	if cfg.StorePath == "" && cfg.StoreType != TypeTSDB {
		if wf.Spec.Provisioning.StorePath != "" && !isTSDBPath(wf.Spec.Provisioning.StorePath) {
			cfg.StorePath = wf.Spec.Provisioning.StorePath
		} else {
			cfg.StorePath = filepath.Join(storeRoot, wf.Namespace, wf.Name+".db")
		}
	}
	b, err := OpenResolved(ctx, c, cfg)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return b, nil
}

// ConfigFromSnapshot builds resolve config from a Snapshot CR.
func ConfigFromSnapshot(snap *kblv1alpha1.Snapshot, storeRoot string) ResolveConfig {
	return ResolveConfig{
		ComputeContextRef: snap.Spec.ComputeContextRef,
		Namespace:         snap.Namespace,
		StoreRoot:         storeRoot,
	}
}

// OpenForSnapshot opens the store for a standalone Snapshot CR.
func OpenForSnapshot(ctx context.Context, c client.Client, snap *kblv1alpha1.Snapshot, storeRoot string) (Backend, error) {
	cfg := ConfigFromSnapshot(snap, storeRoot)
	if cfg.StorePath == "" {
		cfg.StorePath = filepath.Join(storeRoot, snap.Namespace, snap.Name+".db")
	}
	return OpenResolved(ctx, c, cfg)
}

// OpenForDomino opens the shared store for a Domino CR (keyed by snapshot ref).
func OpenForDomino(ctx context.Context, c client.Client, d *kblv1alpha1.Domino, snap *kblv1alpha1.Snapshot, storeRoot string) (Backend, error) {
	if d.Spec.StorePath != "" {
		if isTSDBPath(d.Spec.StorePath) {
			return OpenBackend(Config{Type: TypeTSDB, Endpoint: d.Spec.StorePath})
		}
		return OpenBackend(Config{Type: TypeSQLite, Path: d.Spec.StorePath})
	}
	if snap != nil {
		return OpenForSnapshot(ctx, c, snap, storeRoot)
	}
	root := storeRoot
	if root == "" {
		root = "/var/kbl/store"
	}
	path := filepath.Join(root, d.Namespace, d.Spec.SnapshotRef+".db")
	return OpenBackend(Config{Type: TypeSQLite, Path: path})
}
