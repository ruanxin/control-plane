package update

import (
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process/input"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/reconciler"
	"github.com/sirupsen/logrus"
)

const (
	SCMigrationComponentName = "sc-migration"
)

type SCMigrationStep struct {
	components input.ComponentListProvider
}

type SCMigrationFinalizationStep struct {
	reconcilerClient reconciler.Client
}

func NewSCMigrationStep(components input.ComponentListProvider) *SCMigrationStep {
	return &SCMigrationStep{
		components: components,
	}
}

func NewSCMigrationFinalizationStep(reconcilerClient reconciler.Client) *SCMigrationFinalizationStep {
	return &SCMigrationFinalizationStep{
		reconcilerClient: reconcilerClient,
	}
}

func (s *SCMigrationStep) Name() string {
	return "SCMigration"
}

func (s *SCMigrationStep) Run(operation internal.UpdatingOperation, logger logrus.FieldLogger) (internal.UpdatingOperation, time.Duration, error) {
	for _, c := range operation.LastRuntimeState.ClusterSetup.KymaConfig.Components {
		if c.Component == SCMigrationComponentName {
			// already exists
			return operation, 0, nil
		}
	}
	c, err := getComponentInput(s.components, SCMigrationComponentName, operation.RuntimeVersion)
	if err != nil {
		return operation, 0, err
	}
	operation.LastRuntimeState.ClusterSetup.KymaConfig.Components = append(operation.LastRuntimeState.ClusterSetup.KymaConfig.Components, c)
	operation.RequiresReconcilerUpdate = true
	return operation, 0, nil
}

func (s *SCMigrationFinalizationStep) Name() string {
	return "SCMigrationFinalization"
}

func (s *SCMigrationFinalizationStep) Run(operation internal.UpdatingOperation, logger logrus.FieldLogger) (internal.UpdatingOperation, time.Duration, error) {
	components := make([]reconciler.Component, 0, len(operation.LastRuntimeState.ClusterSetup.KymaConfig.Components))
	for _, c := range operation.LastRuntimeState.ClusterSetup.KymaConfig.Components {
		if c.Component != internal.ServiceCatalogComponentName &&
			c.Component != internal.ServiceCatalogAddonsComponentName &&
			c.Component != internal.HelmBrokerComponentName &&
			c.Component != internal.SCMigrationComponentName &&
			c.Component != internal.ServiceManagerComponentName {
			components = append(components, c)
		} else {
			operation.RequiresReconcilerUpdate = true
		}
	}
	operation.LastRuntimeState.ClusterSetup.KymaConfig.Components = components
	return operation, 0, nil
}
