package services

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-go/models"
	"github.com/smartcontractkit/chainlink-go/models/adapters"
)

func StartJob(run models.JobRun, orm models.ORM) error {
	run.Status = "in progress"
	err := orm.Save(&run)
	if err != nil {
		return runJobError(run, err)
	}

	GetLogger().Infow("Starting job", run.ForLogger()...)
	var prevRun models.TaskRun
	for i, taskRun := range run.TaskRuns {
		prevRun = startTask(taskRun, prevRun.Result)
		run.TaskRuns[i] = prevRun
		err = orm.Save(&run)
		if err != nil {
			return runJobError(run, err)
		}

		GetLogger().Infow("Task finished", run.ForLogger("task", i, "result", prevRun.Result)...)
		if prevRun.Result.Error != nil {
			break
		}
	}

	run.Result = prevRun.Result
	if run.Result.Error != nil {
		run.Status = "errored"
	} else {
		run.Status = "completed"
	}

	GetLogger().Infow("Finished job", run.ForLogger()...)
	return runJobError(run, orm.Save(&run))
}

func startTask(run models.TaskRun, input adapters.RunResult) models.TaskRun {
	run.Status = "in progress"
	adapter, err := run.Adapter()

	if err != nil {
		run.Status = "errored"
		run.Result.Error = err
		return run
	}
	run.Result = adapter.Perform(input)

	if run.Result.Error != nil {
		run.Status = "errored"
	} else {
		run.Status = "completed"
	}

	return run
}

func runJobError(run models.JobRun, err error) error {
	if err != nil {
		return fmt.Errorf("StartJob#%v: %v", run.JobID, err)
	}
	return nil
}