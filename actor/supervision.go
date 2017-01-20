package actor

type Directive int

// Directive determines how a supervisor should handle a failing actor
const (
	// ResumeDirective instructs the supervisor to resume the actor and continue processing messages for the actor
	ResumeDirective Directive = iota

	// RestartDirective instructs the supervisor to restart the actor before processing additional messages
	RestartDirective

	// StopDirective instructs the supervisor to stop the actor
	StopDirective

	// EscalateDirective instructs the supervisor to escalate handling of the failure to the actor's parent
	EscalateDirective
)

type Decider func(child *PID, cause interface{}) Directive

//TODO: as we dont allow remote children or remote SupervisionStrategy
//Instead of letting the parent keep track of child restart stats.
//this info could actually go into each actor, sending it back to the parent as part of the Failure message
type SupervisorStrategy interface {
	HandleFailure(supervisor Supervisor, child *PID, crs *ChildRestartStats, cause interface{})
}

type OneForOneStrategy struct {
	maxNrOfRetries              int
	withinTimeRangeMilliseconds int
	decider                     Decider
}

type Supervisor interface {
	Children() []*PID
	EscalateFailure(who *PID, reason interface{})
}

func (strategy *OneForOneStrategy) HandleFailure(supervisor Supervisor, child *PID, crs *ChildRestartStats, reason interface{}) {
	directive := strategy.decider(child, reason)

	switch directive {
	case ResumeDirective:
		//resume the failing child
		logFailure(child, reason, directive)
		child.sendSystemMessage(resumeMailboxMessage)
	case RestartDirective:
		//try restart the failing child
		if crs.requestRestartPermission(strategy.maxNrOfRetries, strategy.withinTimeRangeMilliseconds) {
			logFailure(child, reason, RestartDirective)
			child.sendSystemMessage(restartMessage)
		} else {
			logFailure(child, reason, StopDirective)
			child.Stop()
		}
	case StopDirective:
		//stop the failing child, no need to involve the crs
		logFailure(child, reason, directive)
		child.Stop()
	case EscalateDirective:
		//send failure to parent
		//supervisor mailbox
		//do not log here, log in the parent handling the error
		supervisor.EscalateFailure(child, reason)
	}
}

func logFailure(child *PID, reason interface{}, directive Directive) {
	event := &SupervisorEvent{
		Child:     child,
		Reason:    reason,
		Directive: directive,
	}
	EventStream.Publish(event)
}

func NewOneForOneStrategy(maxNrOfRetries int, withinTimeRangeMilliseconds int, decider Decider) SupervisorStrategy {
	return &OneForOneStrategy{
		maxNrOfRetries:              maxNrOfRetries,
		withinTimeRangeMilliseconds: withinTimeRangeMilliseconds,
		decider:                     decider,
	}
}

func DefaultDecider(child *PID, reason interface{}) Directive {
	return RestartDirective
}

var defaultSupervisionStrategy = NewOneForOneStrategy(10, 3000, DefaultDecider)

func DefaultSupervisionStrategy() SupervisorStrategy {
	return defaultSupervisionStrategy
}
