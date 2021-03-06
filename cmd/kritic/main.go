package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	color "github.com/fatih/color"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	noColor := flag.Bool("no-color", false, "(Optional) disable ANSI colors")
	watch := flag.Bool("watch", false, "(Optional) loop and show latest values every 5 seconds")
	nodeLabel := flag.String("nodeLabel", "", "(Optional) label and value of node to inspect (e.g. label=value")
	justNodes := flag.Bool("justNodes", false, "(Optional) only show information about nodes")
	showNodeLabels := flag.Bool("showNodeLabels", false, "(Optional) show node labels")
	filterNodeLabels := flag.String("filterNodeLabels", "", "(Optional) filter node labels by this key")
	justTotals := flag.Bool("justTotals", false, "(Optional) only show totals of resource types per node")

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {

		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))

		var splitNodeLabel []string

		if len(*nodeLabel) > 0 {
			splitNodeLabel = strings.Split(*nodeLabel, "=")
		}

		for i := 0; i < len(nodes.Items); i++ {
			node := nodes.Items[i]
			if len(*nodeLabel) > 0 {
				if nodeLabelValue, ok := node.Labels[splitNodeLabel[0]]; ok {
					if nodeLabelValue != splitNodeLabel[1] {
						// don't include this node
						continue
					}
				}
			}

			if *noColor {
				fmt.Printf("Node %s\n", node.Name)
			} else {
				fmt.Printf("Node %s\n", color.HiBlueString(node.Name))
			}

			if *showNodeLabels {
				labelKeys := make([]string, 0)
				for labelKey, _ := range node.Labels {
					labelKeys = append(labelKeys, labelKey)
				}
				sort.Strings(labelKeys)
				for _, k := range labelKeys {
					if len(*filterNodeLabels) > 0 {
						if *filterNodeLabels != k {
							// we are only interested in this particular key
							continue
						}
					}
					if *noColor {
						fmt.Printf("  label %s=%s\n", k, node.Labels[k])
					} else {
						fmt.Printf("  label %s=%s\n", color.HiGreenString(k), color.GreenString(node.Labels[k]))
					}
				}
			}

			if !*justNodes {
				daemonSets := getNodePodsByKind(node, pods.Items, "DaemonSet")
				replicaSets := getNodePodsByKind(node, pods.Items, "ReplicaSet")
				jobs := getNodePodsByKind(node, pods.Items, "Job")

				if *justTotals {
					if *noColor {
						fmt.Printf("daemonSets: %d\n", len(daemonSets))
					} else {
						fmt.Printf("daemonSets: %s\n", color.HiYellowString(fmt.Sprintf("%d", len(daemonSets))))
					}
				} else {
					for j := 0; j < len(daemonSets); j++ {
						daemonSet := daemonSets[j]
						if *noColor {
							fmt.Printf("daemonSet: %s (%s)\n", daemonSet.Name, daemonSet.Status.Phase)
						} else {
							if daemonSet.Status.Phase == "Pending" {
								fmt.Printf("daemonSet: %s (%s)\n", color.YellowString(daemonSet.Name), daemonSet.Status.Phase)
							} else {
								fmt.Printf("daemonSet: %s (%s)\n", color.HiYellowString(daemonSet.Name), daemonSet.Status.Phase)
							}
						}
					}
				}

				if *justTotals {
					if *noColor {
						fmt.Printf("replicaSets: %d\n", len(replicaSets))
					} else {
						fmt.Printf("replicaSets: %s\n", color.HiGreenString(fmt.Sprintf("%d", len(replicaSets))))
					}
				} else {
					for j := 0; j < len(replicaSets); j++ {
						replicaSet := replicaSets[j]
						if *noColor {
							fmt.Printf("replicaSet: %s (%s)\n", replicaSet.Name, replicaSet.Status.Phase)
						} else {
							if replicaSet.Status.Phase == "Pending" {
								fmt.Printf("replicaSet: %s (%s)\n", color.GreenString(replicaSet.Name), replicaSet.Status.Phase)
							} else {
								fmt.Printf("replicaSet: %s (%s)\n", color.HiGreenString(replicaSet.Name), replicaSet.Status.Phase)
							}
						}
					}
				}

				if *justTotals {
					if *noColor {
						fmt.Printf("jobs: %d\n", len(jobs))
					} else {
						fmt.Printf("jobs: %s\n", color.HiCyanString(fmt.Sprintf("%d", len(jobs))))
					}
				} else {
					for j := 0; j < len(jobs); j++ {
						job := jobs[j]
						if *noColor {
							fmt.Printf("job: %s (%s)\n", job.Name, job.Status.Phase)
						} else {
							if job.Status.Phase == "Pending" || job.Status.Phase == "Succeeded" {
								fmt.Printf("job: %s (%s)\n", color.CyanString(job.Name), job.Status.Phase)
							} else {
								fmt.Printf("job: %s (%s)\n", color.HiCyanString(job.Name), job.Status.Phase)
							}
						}
					}
				}
			}
		}

		if *watch {
			time.Sleep(5 * time.Second)
			fmt.Printf("\n")
		} else {
			break
		}
	}
}

func getNodePodsByKind(node v1.Node, pods []v1.Pod, kind string) []v1.Pod {
	results := []v1.Pod{}

	for i := 0; i < len(pods); i++ {
		pod := pods[i]
		if pod.Spec.NodeName == node.Name {
			if len(pod.OwnerReferences) > 0 {
				if pod.OwnerReferences[0].Kind == kind {
					results = append(results, pod)
				}
			}
		}
	}

	return results
}
