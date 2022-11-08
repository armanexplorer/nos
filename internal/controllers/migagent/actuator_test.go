package migagent

import (
	"context"
	"github.com/nebuly-ai/nebulnetes/internal/controllers/migagent/plan"
	"github.com/nebuly-ai/nebulnetes/pkg/gpu"
	"github.com/nebuly-ai/nebulnetes/pkg/gpu/mig"
	"github.com/nebuly-ai/nebulnetes/pkg/resource"
	migtest "github.com/nebuly-ai/nebulnetes/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigActuator_applyDeleteOp(t *testing.T) {
	testCases := []struct {
		name                string
		op                  plan.DeleteOperation
		clientReturnedError gpu.Error

		expectedDeleteCalls      uint
		errorExpected            bool
		atLeastOneDeleteExpected bool
	}{
		{
			name: "Empty delete operation",
			op: plan.DeleteOperation{
				Resources: make(mig.DeviceResourceList, 0),
				Quantity:  0,
			},
			clientReturnedError:      nil,
			expectedDeleteCalls:      0,
			errorExpected:            false,
			atLeastOneDeleteExpected: false,
		},
		{
			name: "Delete op does not have enough candidates",
			op: plan.DeleteOperation{
				Resources: mig.DeviceResourceList{
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-1",
							Status:       resource.StatusUnknown,
						},
						GpuIndex: 0,
					},
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-2",
							Status:       resource.StatusUsed,
						},
						GpuIndex: 0,
					},
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-3",
							Status:       resource.StatusFree,
						},
						GpuIndex: 0,
					},
				},
				Quantity: 2,
			},
			clientReturnedError:      nil,
			expectedDeleteCalls:      1,
			errorExpected:            true,
			atLeastOneDeleteExpected: true,
		},
		{
			name: "More candidates than required, the op should delete only Quantity resources",
			op: plan.DeleteOperation{
				Resources: mig.DeviceResourceList{
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-1",
							Status:       resource.StatusFree,
						},
						GpuIndex: 0,
					},
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-2",
							Status:       resource.StatusFree,
						},
						GpuIndex: 0,
					},
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-3",
							Status:       resource.StatusFree,
						},
						GpuIndex: 0,
					},
				},
				Quantity: 2,
			},
			clientReturnedError:      nil,
			expectedDeleteCalls:      2,
			atLeastOneDeleteExpected: true,
			errorExpected:            false,
		},
		{
			name: "MIG client returns error",
			op: plan.DeleteOperation{
				Resources: mig.DeviceResourceList{
					{
						Device: resource.Device{
							ResourceName: "nvidia.com/mig-1g.10gb",
							DeviceId:     "uid-1",
							Status:       resource.StatusFree,
						},
						GpuIndex: 0,
					},
				},
				Quantity: 1,
			},
			clientReturnedError:      gpu.GenericError.Errorf("an error"),
			expectedDeleteCalls:      1,
			errorExpected:            true,
			atLeastOneDeleteExpected: false,
		},
	}

	var migClient = migtest.MockedMigClient{}
	var actuator = MigActuator{migClient: &migClient}

	for _, tt := range testCases {
		migClient.Reset()
		migClient.ReturnedError = tt.clientReturnedError
		t.Run(tt.name, func(t *testing.T) {
			atLeastOneDelete, err := actuator.applyDeleteOp(context.TODO(), tt.op)
			if tt.errorExpected {
				assert.Error(t, err)
			}
			if !tt.errorExpected {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.atLeastOneDeleteExpected, atLeastOneDelete)
			assert.Equal(t, tt.expectedDeleteCalls, migClient.NumCallsDeleteMigResource)
		})
	}
}

func TestMigActuator_applyCreateOp(t *testing.T) {
	testCases := []struct {
		name                string
		op                  plan.CreateOperation
		clientReturnedError gpu.Error

		expectedCreateCalls      uint
		errorExpected            bool
		atLeastOneCreateExpected bool
	}{
		{
			name: "Empty create operation",
			op: plan.CreateOperation{
				MigProfile: mig.Profile{
					GpuIndex: 0,
					Name:     "1g.10gb",
				},
				Quantity: 0,
			},
			clientReturnedError:      nil,
			expectedCreateCalls:      0,
			errorExpected:            false,
			atLeastOneCreateExpected: false,
		},
		{
			name: "MIG client returns error",
			op: plan.CreateOperation{
				MigProfile: mig.Profile{
					GpuIndex: 0,
					Name:     "1g.10gb",
				},
				Quantity: 1,
			},
			clientReturnedError:      gpu.GenericError.Errorf("an error"),
			expectedCreateCalls:      1,
			errorExpected:            true,
			atLeastOneCreateExpected: false,
		},
		{
			name: "Create success, quantity > 1",
			op: plan.CreateOperation{
				MigProfile: mig.Profile{
					GpuIndex: 0,
					Name:     "1g.10gb",
				},
				Quantity: 4,
			},
			clientReturnedError:      nil,
			expectedCreateCalls:      4,
			errorExpected:            false,
			atLeastOneCreateExpected: true,
		},
	}

	var migClient = migtest.MockedMigClient{}
	var actuator = MigActuator{migClient: &migClient}

	for _, tt := range testCases {
		migClient.Reset()
		migClient.ReturnedError = tt.clientReturnedError
		t.Run(tt.name, func(t *testing.T) {
			atLeastOneCreate, err := actuator.applyCreateOp(context.TODO(), tt.op)
			if tt.errorExpected {
				assert.Error(t, err)
			}
			if !tt.errorExpected {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.atLeastOneCreateExpected, atLeastOneCreate)
			assert.Equal(t, tt.expectedCreateCalls, migClient.NumCallsCreateMigResource)
		})
	}
}
