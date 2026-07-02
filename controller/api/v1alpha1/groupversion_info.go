package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is the group version for KBL API types.
	GroupVersion = schema.GroupVersion{Group: "kbl.io", Version: "v1alpha1"}

	// SchemeBuilder registers KBL types with a scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds KBL types to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
