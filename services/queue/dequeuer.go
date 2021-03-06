package queue

import (
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"time"

	utils "github.com/b-eee/amagi"
	"github.com/globalsign/mgo"
)

const (
	// DequeuerSleepDurationEnv the env var name of the sleep duration
	DequeuerSleepDurationEnv = "QUEUE_DEQUEUER_INTERVAL_MS"

	defaultSleepDuration     = (1 * time.Second)
	defaultMaxConcurrentExec = 1
)

// ExecCallback callback after queueItem.Execute(), original data is passed as arg
type ExecCallback func(Executor) error

// Dequeue loop process for dequeuing the queue
//
// For example:
//
//     go Dequeue("queue_items", callBack, logger, A{}, B{}, C{})
//
func Dequeue(queueCollectionName string, execDelay time.Duration, callback ExecCallback, queueNotificator func(interface{}), loggerFactory func() Logificator, types ...interface{}) {
	QueueCollection = queueCollectionName
	for _, qtype := range types {
		go StartDequeue(qtype, callback, queueNotificator, loggerFactory, execDelay)
	}
}

// StartDequeue main dequeuer
func StartDequeue(qtype interface{}, callback ExecCallback, queueNotificator func(interface{}), loggerFactory func() Logificator, execDelay time.Duration) {
	sleepDuration := getSleepDuration()
	typeName := GetTypeName(qtype)
	gob.RegisterName(typeName, qtype)
	queueItem := Queue{}

	utils.Info(fmt.Sprintf("[Amagi-Queue] Dequeuer started for `%v` with %v sleeping time...", typeName, sleepDuration))

	for {
		func() {
			// TODO: add concurrency settings? like how many max concurrent execution at the same time
			if err := queueItem.Dequeue(typeName, queueNotificator); err != nil {
				if err != mgo.ErrNotFound {
					utils.Info(fmt.Sprintf("[Amagi-Queue] Error during dequeue for `%s`: %v", typeName, err))
				}
				time.Sleep(sleepDuration)
				return
			}
			logger := loggerFactory()
			logger.Initialize(queueItem.ID.Hex())
			defer logger.Finalize()
			defer queueItem.CleanUp()

			itemString := fmt.Sprintf("queue `%v` with Identity `%v`",
				queueItem.ID.Hex(),
				queueItem.ItemExec.Identity(),
			)
			time.Sleep(execDelay)
			defer func() {
				if r := recover(); r != nil {
					utils.Error(fmt.Sprintf("[Amagi-Queue] Queue task panicked: %v", r))
					logger.Error(fmt.Sprintf("Task exited with error: %v", r))
					queueItem.Fail()
				}
			}()
			utils.Info(fmt.Sprintf("[Amagi-Queue] Starting process for %s", itemString))
			procStart := time.Now()
			if err := queueItem.ItemExec.Execute(logger); err != nil {
				utils.Error(fmt.Sprintf("[Amagi-Queue] error queueItem.Execute for %s: %v", itemString, err))
				defer queueItem.Fail()
				return
			}
			if callback != nil {
				if err := callback(queueItem.ItemExec); err != nil {
					utils.Error(fmt.Sprintf("[Amagi-Queue] error queueItem.Execute(callback) for %s: %v", itemString, err))
					defer queueItem.Fail()
					return
				}
			}
			queueItem.Success()
			utils.Info(fmt.Sprintf("[Amagi-Queue] Queued %s is done, took: %v",
				itemString,
				time.Since(procStart),
			))
		}()
	}
}

func getSleepDuration() time.Duration {
	if durationEnv := os.Getenv(DequeuerSleepDurationEnv); durationEnv != "" {
		duration, err := strconv.Atoi(durationEnv)
		if err != nil {
			utils.Error(fmt.Sprintf("[Amagi-Queue] Invalid dequeuer sleep duration value: %v", err))
			utils.Warn(fmt.Sprintf("[Amagi-Queue] Using default sleep duration: %v", defaultSleepDuration))
			return defaultSleepDuration
		}
		return (time.Duration(duration) * time.Millisecond)
	}
	return defaultSleepDuration
}
