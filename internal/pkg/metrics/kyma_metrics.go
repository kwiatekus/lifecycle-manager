package metrics

import (
	"fmt"

	watchermetrics "github.com/kyma-project/runtime-watcher/listener/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
)

const (
	metricKymaState   = "lifecycle_mgr_kyma_state"
	metricModuleState = "lifecycle_mgr_module_state"
	stateLabel        = "state"
	moduleNameLabel   = "module_name"
)

var (
	//nolint:gochecknoglobals
	kymaStateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metricKymaState,
		Help: "Indicates the Status.state for a given Kyma object",
	}, []string{kymaNameLabel, stateLabel, shootIDLabel, instanceIDLabel})
	//nolint:gochecknoglobals
	moduleStateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metricModuleState,
		Help: "Indicates the Status.state for modules of Kyma",
	}, []string{moduleNameLabel, kymaNameLabel, stateLabel, shootIDLabel, instanceIDLabel})
)

func InitKymaMetrics() {
	ctrlmetrics.Registry.MustRegister(kymaStateGauge)
	ctrlmetrics.Registry.MustRegister(moduleStateGauge)
	watchermetrics.Init(ctrlmetrics.Registry)
}

// UpdateAll sets both metrics 'lifecycle_mgr_kyma_state' and 'lifecycle_mgr_module_state' to new states.
func UpdateAll(kyma *v1beta2.Kyma) error {
	shootID, err := ExtractShootID(kyma)
	if err != nil {
		return fmt.Errorf("%w: %w", errMetric, err)
	}
	instanceID, err := ExtractInstanceID(kyma)
	if err != nil {
		return fmt.Errorf("%w: %w", errMetric, err)
	}

	setKymaStateGauge(kyma.Status.State, kyma.Name, shootID, instanceID)
	for _, moduleStatus := range kyma.Status.Modules {
		setModuleStateGauge(moduleStatus.State, moduleStatus.Name, kyma.Name, shootID, instanceID)
	}
	return nil
}

// CleanupMetrics deletes all 'lifecycle_mgr_kyma_state',
// 'lifecycle_mgr_module_state' metrics for the matching Kyma.
func CleanupMetrics(kyma *v1beta2.Kyma) {
	kymaStateGauge.DeletePartialMatch(prometheus.Labels{
		kymaNameLabel: kyma.Name,
	})
	moduleStateGauge.DeletePartialMatch(prometheus.Labels{
		kymaNameLabel: kyma.Name,
	})
}

// RemoveModuleStateMetrics deletes all 'lifecycle_mgr_module_state' metrics for the matching module.
func RemoveModuleStateMetrics(kyma *v1beta2.Kyma, moduleName string) {
	moduleStateGauge.DeletePartialMatch(prometheus.Labels{
		moduleNameLabel: moduleName,
		kymaNameLabel:   kyma.Name,
	})
}

func setKymaStateGauge(newState shared.State, kymaName, shootID, instanceID string) {
	states := shared.AllStates()
	for _, state := range states {
		newValue := calcStateValue(state, newState)
		kymaStateGauge.With(prometheus.Labels{
			kymaNameLabel:   kymaName,
			shootIDLabel:    shootID,
			instanceIDLabel: instanceID,
			stateLabel:      string(state),
		}).Set(newValue)
	}
}

func setModuleStateGauge(newState shared.State, moduleName, kymaName, shootID, instanceID string) {
	states := shared.AllStates()
	for _, state := range states {
		newValue := calcStateValue(state, newState)
		moduleStateGauge.With(prometheus.Labels{
			moduleNameLabel: moduleName,
			kymaNameLabel:   kymaName,
			shootIDLabel:    shootID,
			instanceIDLabel: instanceID,
			stateLabel:      string(state),
		}).Set(newValue)
	}
}

func calcStateValue(state, newState shared.State) float64 {
	if state == newState {
		return 1
	}
	return 0
}