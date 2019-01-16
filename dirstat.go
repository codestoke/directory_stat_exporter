package main

import (
	"fmt"
	"github.com/codestoke/directory_stat_exporter/cfg"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type metricValue struct {
	labels    map[string]string
	name      string
	value     int64
	recursive bool
}

type metric struct {
	metricName   string
	metricHelp   string
	metricType   string
	metricValues map[string]metricValue
}

const (
	namespace              = "dirstat"
	metricFilesInDir       = "files_in_dir"
	metricOldestFileTime   = "oldest_file_time"
	metricCurrentTimestamp = "current_timestamp"
)

var (
	config           cfg.Config
	metricRegister   map[string]metric
	currentTimestamp metric
	cached           bool
	lastRequest      time.Time
)

func main() {
	config = cfg.GetConfig("")

	metricRegister = make(map[string]metric)

	metricRegister[metricFilesInDir] = metric{
		metricName:   metricFilesInDir,
		metricType:   "gauge",
		metricHelp:   "this counts all the files in a directory",
		metricValues: make(map[string]metricValue),
	}
	metricRegister[metricOldestFileTime] = metric{
		metricName:   metricOldestFileTime,
		metricType:   "gauge",
		metricHelp:   "displays the timestamp in unix time of the oldes file",
		metricValues: make(map[string]metricValue),
	}

	lastRequest = time.Unix(0, 0)
	cached = false

	http.HandleFunc("/metrics", handleMetrics)
	if err := http.ListenAndServe(":"+config.ServicePort, nil); err != nil {
		panic(err)
	}
}

func handleMetrics(w http.ResponseWriter, _ *http.Request) {
	if lastRequest.Add(time.Minute*time.Duration(config.CacheTime)).Unix() < time.Now().Unix() {
		// update the cache
		if cached {
			writeMetricsResponse(w)
			go updateMetrics()
		} else {
			updateMetrics()
			writeMetricsResponse(w)
		}
	} else {
		// respond with the cache.
		writeMetricsResponse(w)
	}
}

func writeMetricsResponse(w http.ResponseWriter) {
	_, _ = w.Write([]byte(sprintDirMetric(currentTimestamp)))
	for _, value := range metricRegister {
		_, _ = w.Write([]byte(sprintDirMetric(value)))
	}
}

func updateMetrics() {
	for _, dir := range config.Directories {
		if dir.Recursive {
			metricRegister[metricFilesInDir].metricValues[dir.Path] = metricValue{
				value:     int64(getFileCountInDirRecursively(dir.Path)),
				recursive: dir.Recursive,
				name:      dir.Name,
				labels: map[string]string{
					"dir":       dir.Name,
					"recursive": strconv.FormatBool(dir.Recursive),
				},
			}
			metricRegister[metricOldestFileTime].metricValues[dir.Path] = metricValue{
				value:     int64(getOldestAgeInDirRecursively(dir.Path)),
				recursive: dir.Recursive,
				name:      dir.Name,
				labels: map[string]string{
					"dir":       dir.Name,
					"recursive": strconv.FormatBool(dir.Recursive),
				},
			}
		} else {
			metricRegister[metricFilesInDir].metricValues[dir.Path] = metricValue{
				value:     int64(getFileCountInDir(dir.Path)),
				recursive: dir.Recursive,
				name:      dir.Name,
				labels: map[string]string{
					"dir":       dir.Name,
					"recursive": strconv.FormatBool(dir.Recursive),
				},
			}
			metricRegister[metricOldestFileTime].metricValues[dir.Path] = metricValue{
				value:     int64(getOldestAgeInDir(dir.Path)),
				recursive: dir.Recursive,
				name:      dir.Name,
				labels: map[string]string{
					"dir":       dir.Name,
					"recursive": strconv.FormatBool(dir.Recursive),
				},
			}
		}
	}
	currentTimestamp = metric{
		metricName:   metricCurrentTimestamp,
		metricHelp:   "the current timestamp in unix time.",
		metricType:   "gauge",
		metricValues: map[string]metricValue{"ts": {value: time.Now().Unix()}},
	}

	cached = true
}

// this should be replaced with one more generic generator.
//func sprintCurrentTimestamp(m metric) string {
//	str := ""
//	str += fmt.Sprintf("# HELP %s_%s %s\n", namespace, m.metricName, m.metricHelp)
//	str += fmt.Sprintf("# TYPE %s_%s %s\n", namespace, m.metricName, m.metricType)
//	for _, v := range m.metricValues {
//		str += fmt.Sprintf("%s_%s %v\n", namespace, m.metricName, v.value)
//	}
//	return str
//}

func sprintDirMetric(m metric) string {
	str := ""
	str += fmt.Sprintf("# HELP %s_%s %s\n", namespace, m.metricName, m.metricHelp)
	str += fmt.Sprintf("# TYPE %s_%s %s\n", namespace, m.metricName, m.metricType)
	for _, v := range m.metricValues {
		//str += fmt.Sprintf("%s_%s{dir=\"%s\",recursive=\"%t\"} %v\n", namespace, m.metricName, v.name, v.recursive, v.value)
		str += sprintMetric(namespace, m.metricName, v.value, v.labels)
	}
	return str
}

func sprintMetric(ns string, name string, value int64, labels map[string]string) string {
	strLbls := ""
	if labels != nil {
		var lblArr []string
		strLbls += "{"
		for k, v := range labels {
			lblArr = append(lblArr, fmt.Sprintf("%s=\"%s\"", k, v))
		}
		strLbls += strings.Join(lblArr, ",")
		strLbls += "}"
	}
	str := fmt.Sprintf("%s_%s%s %v\n", ns, name, strLbls, value)
	return str
}
