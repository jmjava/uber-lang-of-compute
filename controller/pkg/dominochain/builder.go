package dominochain

import (
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

const (
	HandoffMountPath  = "/kbl/handoff"
	SnapshotMountPath = "/kbl/input"
	HandoffVolumeName = "handoff"
	SnapshotVolumeName = "snapshot"
	PlaceholderImage  = "registry.k8s.io/pause:3.9"
	DefaultRunnerImage = "ghcr.io/jmjava/kbl-domino-runner:latest"
	// DefaultJuliaRunnerImage includes Julia + pre-instantiated controller/julia project (Phase 20).
	DefaultJuliaRunnerImage = "ghcr.io/jmjava/kbl-domino-runner-julia:latest"
	// JuliaProjectContainerPath is KBL_JULIA_PROJECT inside the Julia runner image.
	JuliaProjectContainerPath = "/opt/kbl/julia"
	LabelManagedBy    = "app.kubernetes.io/managed-by"
	LabelDominoChain  = "kbl.io/dominochain"
)

// Builder constructs Kubernetes resources for domino chains.
type Builder struct {
	RunnerImage string
}

func (b *Builder) runnerImage(chain *kblv1alpha1.DominoChain) string {
	if chain.Spec.RunnerImage != "" {
		return chain.Spec.RunnerImage
	}
	if b != nil && b.RunnerImage != "" {
		return b.RunnerImage
	}
	return DefaultRunnerImage
}

func (b *Builder) stepImage(chain *kblv1alpha1.DominoChain, step kblv1alpha1.DominoStepSpec) string {
	if step.Image != "" {
		return step.Image
	}
	return b.runnerImage(chain)
}

// SnapshotConfigMap builds a ConfigMap holding sealed snapshot JSON.
func (b *Builder) SnapshotConfigMap(chain *kblv1alpha1.DominoChain, snapshotJSON string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chain.Name + "-snapshot",
			Namespace: chain.Namespace,
			Labels:    chainLabels(chain.Name),
		},
		Data: map[string]string{
			"snapshot.json": snapshotJSON,
		},
	}
}

// BuildInitChainPod returns a Pod whose initContainers run the domino chain sequentially.
func (b *Builder) BuildInitChainPod(chain *kblv1alpha1.DominoChain) *corev1.Pod {
	podName := chain.Name + "-chain"
	volumes := []corev1.Volume{
		{Name: HandoffVolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		{Name: SnapshotVolumeName, VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: chain.Name + "-snapshot"},
			},
		}},
	}

	initContainers := make([]corev1.Container, len(chain.Spec.Steps))
	for i, step := range chain.Spec.Steps {
		inputPath := filepath.Join(SnapshotMountPath, "snapshot.json")
		if i > 0 {
			inputPath = filepath.Join(HandoffMountPath, "output.json")
		}
		initContainers[i] = corev1.Container{
			Name:            slotName(i, step.Name),
			Image:           b.stepImage(chain, step),
			Command:         b.stepCommand(step),
			Env:             b.stepEnv(step, inputPath),
			VolumeMounts:    handoffMounts(),
			ImagePullPolicy: corev1.PullIfNotPresent,
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: chain.Namespace,
			Labels:    chainLabels(chain.Name),
			Annotations: map[string]string{
				"kbl.io/runtime": string(chain.Spec.Runtime),
			},
		},
		Spec: corev1.PodSpec{
			InitContainers: initContainers,
			Containers: []corev1.Container{{
				Name:            "chain-complete",
				Image:           PlaceholderImage,
				Command:         []string{"/pause"},
				VolumeMounts:    []corev1.VolumeMount{{Name: HandoffVolumeName, MountPath: HandoffMountPath}},
				ImagePullPolicy: corev1.PullIfNotPresent,
			}},
			Volumes:                       volumes,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: ptrInt64(30),
		},
	}
	if chain.Spec.NodeSelector != nil {
		pod.Spec.NodeSelector = chain.Spec.NodeSelector
	}
	return pod
}

// BuildOpenKruisePod returns a Pod with placeholder containers for hot-swap slots.
func (b *Builder) BuildOpenKruisePod(chain *kblv1alpha1.DominoChain) *corev1.Pod {
	podName := chain.Name + "-chain"
	volumes := []corev1.Volume{
		{Name: HandoffVolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		{Name: SnapshotVolumeName, VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: chain.Name + "-snapshot"},
			},
		}},
	}

	containers := make([]corev1.Container, len(chain.Spec.Steps))
	for i, step := range chain.Spec.Steps {
		containers[i] = corev1.Container{
			Name:            slotName(i, step.Name),
			Image:           PlaceholderImage,
			Command:         []string{"/pause"},
			VolumeMounts:    handoffMounts(),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{
				{Name: "KBL_STEP_NAME", Value: step.Name},
				{Name: "KBL_STEP_INDEX", Value: fmt.Sprintf("%d", i)},
			},
		}
		_ = step // real image applied via ContainerRecreateRequest
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: chain.Namespace,
			Labels:    chainLabels(chain.Name),
			Annotations: map[string]string{
				"kbl.io/runtime":       string(kblv1alpha1.DominoChainRuntimeOpenKruise),
				"kbl.io/runner-image":  b.runnerImage(chain),
				"kbl.io/active-slots":  "2",
			},
		},
		Spec: corev1.PodSpec{
			Containers:                    containers,
			Volumes:                       volumes,
			RestartPolicy:                 corev1.RestartPolicyNever,
			TerminationGracePeriodSeconds: ptrInt64(30),
		},
	}
	if chain.Spec.NodeSelector != nil {
		pod.Spec.NodeSelector = chain.Spec.NodeSelector
	}
	return pod
}

func (b *Builder) stepCommand(step kblv1alpha1.DominoStepSpec) []string {
	if len(step.Args) > 0 {
		return step.Args
	}
	return nil // use image entrypoint (domino-runner)
}

func (b *Builder) stepEnv(step kblv1alpha1.DominoStepSpec, inputPath string) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{Name: "KBL_COMMAND", Value: step.Command},
		{Name: "KBL_INPUT", Value: inputPath},
		{Name: "KBL_OUTPUT", Value: filepath.Join(HandoffMountPath, "output.json")},
		{Name: "KBL_STEP_NAME", Value: step.Name},
	}
	if IsJuliaCommand(step.Command) {
		env = append(env,
			corev1.EnvVar{Name: "KBL_JULIA_PROJECT", Value: JuliaProjectContainerPath},
			corev1.EnvVar{Name: "KBL_JULIA_BIN", Value: "julia"},
		)
	}
	return env
}

// IsJuliaCommand reports whether a domino command dispatches to the Julia executor.
func IsJuliaCommand(command string) bool {
	return strings.HasPrefix(command, "julia:")
}

func handoffMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: HandoffVolumeName, MountPath: HandoffMountPath},
		{Name: SnapshotVolumeName, MountPath: SnapshotMountPath, ReadOnly: true},
	}
}

func slotName(index int, stepName string) string {
	return fmt.Sprintf("slot-%d-%s", index, sanitize(stepName))
}

func sanitize(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name) && len(out) < 20; i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			out = append(out, c)
		} else if c >= 'A' && c <= 'Z' {
			out = append(out, c+32)
		} else if c == '_' {
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "step"
	}
	return string(out)
}

func chainLabels(chainName string) map[string]string {
	return map[string]string{
		LabelDominoChain: chainName,
		LabelManagedBy:   "kbl-controller",
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}

// StepContainerName resolves the container name for a step index.
func StepContainerName(chain *kblv1alpha1.DominoChain, index int) string {
	if index < 0 || index >= len(chain.Spec.Steps) {
		return ""
	}
	return slotName(index, chain.Spec.Steps[index].Name)
}
