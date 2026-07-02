package controller

import (
	"context"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func publishWorkflowCDC(ctx context.Context, mv *kblv1alpha1.Multiverse, wf *kblv1alpha1.Workflow, result *types.RunResult) {
	if mv == nil || !mv.Spec.Sync.Enabled || result == nil {
		return
	}

	envs := cdc.ExportFromWorkflow(wf, result)
	if len(envs) == 0 {
		return
	}

	pub := cdcPublisherForMultiverse(mv)
	_ = pub.PublishBatch(ctx, result.SnapshotID, envs)
}

func cdcPublisherForMultiverse(mv *kblv1alpha1.Multiverse) cdc.Publisher {
	if mv.Spec.Sync != nil && mv.Spec.Sync.Kafka != nil && len(mv.Spec.Sync.Kafka.Brokers) > 0 {
		topic := mv.Spec.Sync.Kafka.CDCTopic
		if topic == "" {
			topic = cdc.DefaultTopic
		}
		pub, err := cdc.NewKafkaPublisher(cdc.KafkaConfig{
			Brokers: mv.Spec.Sync.Kafka.Brokers,
			Topic:   topic,
		})
		if err == nil {
			return pub
		}
	}
	return cdc.MemoryPublisher{Bus: cdc.DefaultMemory()}
}

func cdcConsumerForReplica(rr *kblv1alpha1.ReadReplica) cdc.Consumer {
	if rr.Spec.CDCSync != nil && len(rr.Spec.CDCSync.Brokers) > 0 {
		topic := rr.Spec.CDCSync.Topic
		if topic == "" {
			topic = cdc.DefaultTopic
		}
		group := rr.Spec.CDCSync.GroupID
		if group == "" {
			group = "kbl-cdc-" + rr.Name
		}
		cons, err := cdc.NewKafkaConsumer(cdc.KafkaConfig{
			Brokers: rr.Spec.CDCSync.Brokers,
			Topic:   topic,
			GroupID: group,
		})
		if err == nil {
			return cons
		}
	}
	return cdc.MemoryConsumer{Bus: cdc.DefaultMemory()}
}

func cdcSyncSpecFromMultiverse(mv *kblv1alpha1.Multiverse) *kblv1alpha1.CDCSyncSpec {
	if mv.Spec.Sync == nil || mv.Spec.Sync.Kafka == nil {
		return &kblv1alpha1.CDCSyncSpec{Topic: cdc.DefaultTopic}
	}
	topic := mv.Spec.Sync.Kafka.CDCTopic
	if topic == "" {
		topic = cdc.DefaultTopic
	}
	return &kblv1alpha1.CDCSyncSpec{
		Brokers: mv.Spec.Sync.Kafka.Brokers,
		Topic:   topic,
		GroupID: mv.Spec.Sync.Kafka.GroupID,
	}
}
