package store

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

func TestNodeStore(t *testing.T) {
	cases := []generateMetricsTestCase{
		// Verify standard node role labels
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
				kube_node_role{node="node-1",role="control-plane"} 1
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify alternative role labels - kubernetes.io/role
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
					Labels: map[string]string{
						"kubernetes.io/role": "worker",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
				kube_node_role{node="node-2",role="worker"} 1
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify alternative role labels - simple role label
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
					Labels: map[string]string{
						"role": "infra",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
				kube_node_role{node="node-3",role="infra"} 1
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify multiple role labels - should prefer standard Kubernetes role
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-4",
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
						"kubernetes.io/role": "master",
						"role": "control-plane",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
				kube_node_role{node="node-4",role="control-plane"} 1
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify behavior with empty role values
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-5",
					Labels: map[string]string{
						"role": "",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify populating base metrics
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KernelVersion:           "kernel",
						KubeletVersion:          "kubelet",
						KubeProxyVersion:        "kubeproxy",
						OSImage:                 "osimage",
						ContainerRuntimeVersion: "rkt",
						SystemUUID:              "6a934e21-5207-4a84-baea-3a952d926c80",
					},
					Addresses: []v1.NodeAddress{
						{Type: "InternalIP", Address: "1.2.3.4"},
					},
				},
				Spec: v1.NodeSpec{
					ProviderID: "provider://i-uniqueid",
					PodCIDR:    "172.24.10.0/24",
				},
			},
			Want: `
				# HELP kube_node_info [STABLE] Information about a cluster node.
				# HELP kube_node_labels [STABLE] Kubernetes labels converted to Prometheus labels.
				# HELP kube_node_spec_unschedulable [STABLE] Whether a node can schedule new pods.
				# TYPE kube_node_info gauge
				# TYPE kube_node_labels gauge
				# TYPE kube_node_spec_unschedulable gauge
				kube_node_info{container_runtime_version="rkt",kernel_version="kernel",kubelet_version="kubelet",kubeproxy_version="deprecated",node="127.0.0.1",os_image="osimage",pod_cidr="172.24.10.0/24",provider_id="provider://i-uniqueid",internal_ip="1.2.3.4",system_uuid="6a934e21-5207-4a84-baea-3a952d926c80"} 1
				kube_node_spec_unschedulable{node="127.0.0.1"} 0
			`,
			MetricNames: []string{"kube_node_spec_unschedulable", "kube_node_labels", "kube_node_info"},
		},
		// Verify resource metrics
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "127.0.0.1",
					CreationTimestamp: metav1.Time{Time: time.Unix(1500000000, 0)},
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
				},
				Spec: v1.NodeSpec{
					Unschedulable: true,
					ProviderID:    "provider://i-randomidentifier",
					PodCIDR:       "172.24.10.0/24",
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KernelVersion:           "kernel",
						KubeletVersion:          "kubelet",
						KubeProxyVersion:        "kubeproxy",
						OSImage:                 "osimage",
						ContainerRuntimeVersion: "rkt",
						SystemUUID:              "6a934e21-5207-4a84-baea-3a952d926c80",
					},
					Addresses: []v1.NodeAddress{
						{Type: "InternalIP", Address: "1.2.3.4"},
					},
					Capacity: v1.ResourceList{
						v1.ResourceCPU:                    resource.MustParse("4.3"),
						v1.ResourceMemory:                 resource.MustParse("2G"),
						v1.ResourcePods:                   resource.MustParse("1000"),
						v1.ResourceStorage:                resource.MustParse("3G"),
						v1.ResourceEphemeralStorage:       resource.MustParse("4G"),
						v1.ResourceName("nvidia.com/gpu"): resource.MustParse("4"),
					},
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:                    resource.MustParse("3"),
						v1.ResourceMemory:                 resource.MustParse("1G"),
						v1.ResourcePods:                   resource.MustParse("555"),
						v1.ResourceStorage:                resource.MustParse("2G"),
						v1.ResourceEphemeralStorage:       resource.MustParse("3G"),
						v1.ResourceName("nvidia.com/gpu"): resource.MustParse("1"),
					},
				},
			},
			Want: `
				# HELP kube_node_created [STABLE] Unix creation timestamp
				# HELP kube_node_info [STABLE] Information about a cluster node.
				# HELP kube_node_labels [STABLE] Kubernetes labels converted to Prometheus labels.
				# HELP kube_node_role The role of a cluster node.
				# HELP kube_node_spec_unschedulable [STABLE] Whether a node can schedule new pods.
				# HELP kube_node_status_allocatable [STABLE] The allocatable for different resources of a node that are available for scheduling.
				# HELP kube_node_status_capacity [STABLE] The capacity for different resources of a node.
				# TYPE kube_node_created gauge
				# TYPE kube_node_info gauge
				# TYPE kube_node_labels gauge
				# TYPE kube_node_role gauge
				# TYPE kube_node_spec_unschedulable gauge
				# TYPE kube_node_status_allocatable gauge
				# TYPE kube_node_status_capacity gauge
				kube_node_created{node="127.0.0.1"} 1.5e+09
				kube_node_info{container_runtime_version="rkt",kernel_version="kernel",kubelet_version="kubelet",kubeproxy_version="deprecated",node="127.0.0.1",os_image="osimage",pod_cidr="172.24.10.0/24",provider_id="provider://i-randomidentifier",internal_ip="1.2.3.4",system_uuid="6a934e21-5207-4a84-baea-3a952d926c80"} 1
				kube_node_role{node="127.0.0.1",role="master"} 1
				kube_node_spec_unschedulable{node="127.0.0.1"} 1
				kube_node_status_allocatable{node="127.0.0.1",resource="cpu",unit="core"} 3
				kube_node_status_allocatable{node="127.0.0.1",resource="ephemeral_storage",unit="byte"} 3e+09
				kube_node_status_allocatable{node="127.0.0.1",resource="memory",unit="byte"} 1e+09
				kube_node_status_allocatable{node="127.0.0.1",resource="nvidia_com_gpu",unit="integer"} 1
				kube_node_status_allocatable{node="127.0.0.1",resource="pods",unit="integer"} 555
				kube_node_status_allocatable{node="127.0.0.1",resource="storage",unit="byte"} 2e+09
				kube_node_status_capacity{node="127.0.0.1",resource="cpu",unit="core"} 4.3
				kube_node_status_capacity{node="127.0.0.1",resource="ephemeral_storage",unit="byte"} 4e+09
				kube_node_status_capacity{node="127.0.0.1",resource="memory",unit="byte"} 2e+09
				kube_node_status_capacity{node="127.0.0.1",resource="nvidia_com_gpu",unit="integer"} 4
				kube_node_status_capacity{node="127.0.0.1",resource="pods",unit="integer"} 1000
				kube_node_status_capacity{node="127.0.0.1",resource="storage",unit="byte"} 3e+09
			`,
			MetricNames: []string{
				"kube_node_status_capacity",
				"kube_node_status_allocatable",
				"kube_node_spec_unschedulable",
				"kube_node_labels",
				"kube_node_role",
				"kube_node_info",
				"kube_node_created",
			},
		},
		// Verify node conditions
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{Type: v1.NodeNetworkUnavailable, Status: v1.ConditionTrue},
						{Type: v1.NodeReady, Status: v1.ConditionTrue},
						{Type: v1.NodeConditionType("CustomizedType"), Status: v1.ConditionTrue},
					},
				},
			},
			Want: `
				# HELP kube_node_status_condition [STABLE] The condition of a cluster node.
				# TYPE kube_node_status_condition gauge
				kube_node_status_condition{condition="CustomizedType",node="127.0.0.1",status="false"} 0
				kube_node_status_condition{condition="CustomizedType",node="127.0.0.1",status="true"} 1
				kube_node_status_condition{condition="CustomizedType",node="127.0.0.1",status="unknown"} 0
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="false"} 0
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="true"} 1
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="unknown"} 0
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="false"} 0
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="true"} 1
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="unknown"} 0
			`,
			MetricNames: []string{"kube_node_status_condition"},
		},
		// Verify node taints
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
				},
				Spec: v1.NodeSpec{
					Taints: []v1.Taint{
						{Key: "node.kubernetes.io/memory-pressure", Value: "true", Effect: v1.TaintEffectPreferNoSchedule},
						{Key: "node.kubernetes.io/disk-pressure", Value: "true", Effect: v1.TaintEffectNoSchedule},
						{Key: "dedicated", Effect: v1.TaintEffectPreferNoSchedule},
					},
				},
			},
			Want: `
				# HELP kube_node_spec_taint [STABLE] The taint of a cluster node.
				# TYPE kube_node_spec_taint gauge
				kube_node_spec_taint{effect="PreferNoSchedule",key="node.kubernetes.io/memory-pressure",node="127.0.0.1",value="true"} 1
				kube_node_spec_taint{effect="NoSchedule",key="node.kubernetes.io/disk-pressure",node="127.0.0.1",value="true"} 1
				kube_node_spec_taint{effect="PreferNoSchedule",key="dedicated",node="127.0.0.1",value=""} 1
			`,
			MetricNames: []string{"kube_node_spec_taint"},
		},
		// Verify node address information
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{Type: "InternalIP", Address: "1.2.3.4"},
						{Type: "InternalIP", Address: "fc00::"},
						{Type: "ExternalIP", Address: "5.6.7.8"},
						{Type: "ExternalIP", Address: "2001:db8::"},
						{Type: "Hostname", Address: "node1.example.com"},
					},
				},
			},
			Want: `
				# HELP kube_node_status_addresses Node address information.
				# TYPE kube_node_status_addresses gauge
				kube_node_status_addresses{node="127.0.0.1",type="InternalIP",address="1.2.3.4"} 1
				kube_node_status_addresses{node="127.0.0.1",type="InternalIP",address="fc00::"} 1
				kube_node_status_addresses{node="127.0.0.1",type="ExternalIP",address="5.6.7.8"} 1
				kube_node_status_addresses{node="127.0.0.1",type="ExternalIP",address="2001:db8::"} 1
				kube_node_status_addresses{node="127.0.0.1",type="Hostname",address="node1.example.com"} 1
			`,
			MetricNames: []string{"kube_node_status_addresses"},
		},
		// Verify unset fields are skipped
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Status:     v1.NodeStatus{},
				Spec:       v1.NodeSpec{},
			},
			Want: `
				# HELP kube_node_info [STABLE] Information about a cluster node.
				# TYPE kube_node_info gauge
				kube_node_info{container_runtime_version="",kernel_version="",kubelet_version="",kubeproxy_version="deprecated",node="",os_image="",pod_cidr="",provider_id="",internal_ip="",system_uuid=""} 1
			`,
			MetricNames: []string{"kube_node_info"},
		},
		// Verify node conditions with unknown status
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{Type: v1.NodeNetworkUnavailable, Status: v1.ConditionUnknown},
						{Type: v1.NodeReady, Status: v1.ConditionUnknown},
					},
				},
			},
			Want: `
				# HELP kube_node_status_condition [STABLE] The condition of a cluster node.
				# TYPE kube_node_status_condition gauge
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="false"} 0
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="true"} 0
				kube_node_status_condition{condition="NetworkUnavailable",node="127.0.0.1",status="unknown"} 1
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="false"} 0
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="true"} 0
				kube_node_status_condition{condition="Ready",node="127.0.0.1",status="unknown"} 1
			`,
			MetricNames: []string{"kube_node_status_condition"},
		},
		// Verify multiple standard node role labels
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "127.0.0.1",
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
						"node-role.kubernetes.io/master":        "",
					},
				},
			},
			Want: `
				# HELP kube_node_role The role of a cluster node.
				# TYPE kube_node_role gauge
				kube_node_role{node="127.0.0.1",role="control-plane"} 1
				kube_node_role{node="127.0.0.1",role="master"} 1
			`,
			MetricNames: []string{"kube_node_role"},
		},
		// Verify node deletion timestamp
		{
			Obj: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "127.0.0.1",
					DeletionTimestamp: &metav1.Time{Time: time.Unix(1500000000, 0)},
				},
			},
			Want: `
				# HELP kube_node_deletion_timestamp Unix deletion timestamp
				# TYPE kube_node_deletion_timestamp gauge
				kube_node_deletion_timestamp{node="127.0.0.1"} 1.5e+09
			`,
			MetricNames: []string{"kube_node_deletion_timestamp"},
		},
	}

	for i, c := range cases {
		c.Func = generator.ComposeMetricGenFuncs(nodeMetricFamilies(nil, nil))
		c.Headers = generator.ExtractMetricFamilyHeaders(nodeMetricFamilies(nil, nil))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
