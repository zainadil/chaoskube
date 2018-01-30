package main

import (
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/linki/chaoskube/chaoskube"
	"github.com/linki/chaoskube/util"
)

var (
	labelString      string
	annString        string
	nsString         string
	excludedWeekdays string
	excludedHours    string
	timezone         string
	master           string
	kubeconfig       string
	interval         time.Duration
	inCluster        bool
	dryRun           bool
	debug            bool
	version          string
)

func init() {
	kingpin.Flag("labels", "A set of labels to restrict the list of affected pods. Defaults to everything.").StringVar(&labelString)
	kingpin.Flag("annotations", "A set of annotations to restrict the list of affected pods. Defaults to everything.").StringVar(&annString)
	kingpin.Flag("namespaces", "A set of namespaces to restrict the list of affected pods. Defaults to everything.").StringVar(&nsString)
	kingpin.Flag("excluded-weekdays", "A list of weekdays when termination is suspended, e.g. sat,sun").StringVar(&excludedWeekdays)
	kingpin.Flag("excluded-hours", "TODO, e.g. 00-08,12-13,16-24").StringVar(&excludedHours)
	kingpin.Flag("timezone", "The timezone to apply when detecting the current weekday, e.g. UTC, Local, Europe/Berlin. Defaults to UTC.").Default("UTC").StringVar(&timezone)
	kingpin.Flag("master", "The address of the Kubernetes cluster to target").StringVar(&master)
	kingpin.Flag("kubeconfig", "Path to a kubeconfig file").StringVar(&kubeconfig)
	kingpin.Flag("interval", "Interval between Pod terminations").Default("10m").DurationVar(&interval)
	kingpin.Flag("dry-run", "If true, don't actually do anything.").Default("true").BoolVar(&dryRun)
	kingpin.Flag("debug", "Enable debug logging.").BoolVar(&debug)
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if dryRun {
		log.Infof("Dry run enabled. I won't kill anything. Use --no-dry-run when you're ready.")
	}

	client, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	labelSelector, err := labels.Parse(labelString)
	if err != nil {
		log.Fatal(err)
	}

	annotations, err := labels.Parse(annString)
	if err != nil {
		log.Fatal(err)
	}

	namespaces, err := labels.Parse(nsString)
	if err != nil {
		log.Fatal(err)
	}

	if !labelSelector.Empty() {
		log.Infof("Filtering pods by labels: %s", labelSelector.String())
	}

	if !annotations.Empty() {
		log.Infof("Filtering pods by annotations: %s", annotations.String())
	}

	if !namespaces.Empty() {
		log.Infof("Filtering pods by namespaces: %s", namespaces.String())
	}

	parsedTimezone, err := time.LoadLocation(timezone)
	if err != nil {
		log.Fatal(err)
	}
	timezoneName, _ := time.Now().In(parsedTimezone).Zone()
	log.Infof("Using time zone: %s (%s)", parsedTimezone.String(), timezoneName)

	parsedWeekdays := util.ParseWeekdays(excludedWeekdays)
	if len(parsedWeekdays) > 0 {
		log.Infof("Excluding weekdays: %s", parsedWeekdays)
	}

	parsedHours := parseHours(excludedHours)
	if len(parsedHours) > 0 {
		log.Infof("Excluding hours: %s", parsedHours)
	}

	spew.Dump(parsedHours)

	chaoskube := chaoskube.New(
		client,
		labelSelector,
		annotations,
		namespaces,
		parsedWeekdays,
		[]chaoskube.TimePeriod{},
		parsedTimezone,
		log.StandardLogger(),
		dryRun,
		time.Now().UTC().UnixNano(),
	)

	for {
		if err := chaoskube.TerminateVictim(); err != nil {
			log.Fatal(err)
		}

		log.Debugf("Sleeping for %s...", interval)
		time.Sleep(interval)
	}
}

func parseHours(hours string) []chaoskube.TimePeriod {
	parsedHours := []chaoskube.TimePeriod{}
	for _, tp := range strings.Split(hours, ",") {
		t := strings.Split(tp, "-")
		if len(t) != 2 {
			log.Fatal("wrong format")
		}

		t1, err := time.ParseInLocation(time.Kitchen, t[0], time.Local)
		if err != nil {
			log.Fatal(err)
		}

		t2, err := time.ParseInLocation(time.Kitchen, t[1], time.Local)
		if err != nil {
			log.Fatal(err)
		}

		year, month, day := time.Now().Date()
		t1 = t1.AddDate(year, int(month), day).Local()
		t2 = t2.AddDate(year, int(month), day).Local()

		parsedHours = append(parsedHours, chaoskube.TimePeriod{
			From: t1,
			To:   t2,
		})
	}
	return parsedHours
}

func newClient() (*kubernetes.Clientset, error) {
	if kubeconfig == "" {
		if _, err := os.Stat(clientcmd.RecommendedHomeFile); err == nil {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
	}

	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return nil, err
	}

	log.Infof("Targeting cluster at %s", config.Host)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
