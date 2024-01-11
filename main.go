package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace       = flag.String("n", "", "namespace")
	container       = flag.String("c", ".*", "container_regexp")
	containerRegexp *regexp.Regexp
	source          = flag.String("s", ".*", "source_regexp")
	sourceRegexp    *regexp.Regexp
	since           = flag.Duration("since", 0, "since")
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run(ctx context.Context) (err error) {
	flag.Parse()

	if *namespace == "" {
		flag.Usage()
		return fmt.Errorf("-n: required")
	}

	containerRegexp, err = regexp.Compile(*container)
	if err != nil {
		flag.Usage()
		return fmt.Errorf("-c %s: %w", *container, err)
	}

	sourceRegexp, err = regexp.Compile(*source)
	if err != nil {
		flag.Usage()
		return fmt.Errorf("-s %s: %w", *source, err)
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return err
	}

	client, err := corev1client.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := client.Pods(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, p := range pods.Items {
		for _, c := range p.Spec.Containers {
			if !containerRegexp.MatchString(c.Name) {
				continue
			}

			opts := &corev1.PodLogOptions{
				Container:  c.Name,
				Timestamps: true,
			}

			if *since != 0 {
				sinceSeconds := int64(*since / time.Second)
				opts.SinceSeconds = &sinceSeconds
			}

			rc, err := client.Pods(*namespace).GetLogs(p.Name, opts).Stream(ctx)
			if err != nil {
				return err
			}
			defer rc.Close()

			r := bufio.NewReader(rc)

			tables := map[string]*table{}

			for {
				line, err := r.ReadBytes('\n')
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}

				timestamp, line, _ := bytes.Cut(line, []byte{' '})

				var o map[string]any
				err = json.Unmarshal(line, &o)
				if err != nil {
					o = map[string]any{"msg": string(line)}
				}

				o["timestamp"] = string(timestamp)

				source, _ := o["source"].(string)
				if !sourceRegexp.MatchString(source) {
					continue
				}

				if tables[source] == nil {
					tables[source] = newTable()
				}

				tables[source].addRow(o)
			}

			sources := lo.Keys(tables)
			sort.Strings(sources)

			for _, source := range sources {
				fmt.Printf("namespace=%s, pod=%s, container=%s, source=%s\n\n", *namespace, p.Name, c.Name, source)
				tables[source].print()
			}
		}
	}

	return nil
}
