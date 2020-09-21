package operator

import (
	// "net/http"
	"testing"
	"github.com/stretchr/testify/suite"

	"reflect"
	"sync"
	"encoding/json"

	"google.golang.org/grpc"
	"go.uber.org/zap"

	"harmonia.com/steward/operator/aggregateServer"
	"harmonia.com/steward/operator/util"
)

type AggregateServerOperatorTestSuite struct{
	suite.Suite
}

func TestAggregateServerOperator(t *testing.T) {
	suite.Run(t, new(AggregateServerOperatorTestSuite))
}

func (suite *AggregateServerOperatorTestSuite) SetupSuite() {
	rawJSON := []byte(`{
		"level": "debug",
	  "encoding": "json",
	  "outputPaths": ["/dev/null"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)
  
  var cfg zap.Config
  if err := json.Unmarshal(rawJSON, &cfg); err != nil {
	  panic(err)
  }
  logger, err := cfg.Build()
  if err != nil {
	  panic(err)
  }
  defer logger.Sync()
  
  logger.Info("logger construction succeeded")
}

func (suite *AggregateServerOperatorTestSuite) BeforeTest(_, _ string) {
	// suite.T().Log("before Test")
	// suite.T().Log(suite.aggOp)
}

func (suite *AggregateServerOperatorTestSuite) TestNew() {
	aggOp := NewOperator["aggregator"](
		"appGrpcServerURI", 
		"url1",
		"url2",
		"",
		[]string{"url3", "url4"},
		func() {},
	)
	suite.Assert().Equal(
		"url3",
		aggOp.payload.(aggregateServer.Payload).EdgeModelRepoGitHttpURLs[0],
	)
	suite.Assert().Equal(
		"url4",
		aggOp.payload.(aggregateServer.Payload).EdgeModelRepoGitHttpURLs[1],
	)
}

func (suite *AggregateServerOperatorTestSuite) TestDispatch() {
	type fromState struct {}
	type toState struct {}
	type invalidState struct {}
	//---------
	type action struct {}
	type invalidAction struct {}
	//---------

	dummyWebhookToAction := func(_ *util.Webhook, _ util.AbstractOperator) (util.Action, error) {
		return nil, nil
	}
	dummyPullRepoNotification := func(_ util.AbstractOperator) ([]util.Action, error) {
		return nil, nil

	}
	dummyGrpcServerRegister := func(_ *grpc.Server, _ util.AbstractOperator) {
		return
	}

	var testChan chan string

	type testCase struct {
		fromState util.State
		action	util.Action
		stateTransit util.StateTransit
		testThen bool
		// Expected
		toState util.State
		afterThen func(*Operator, testCase) bool
	}
	testCases := map[string] testCase {
		"success": {
			fromState{},
			action{},
			util.StateTransit {
				reflect.TypeOf(fromState{}): {
					reflect.TypeOf(action{}): func(state util.State, action util.Action, operator util.AbstractOperator) (util.State, []func()) {
						return toState{}, []func() {
							func() {
								_ = <-testChan
								operator.(*Operator).payload = struct {
									foo string
								} {"bar"}
								testChan <- "finish"
							},
						}
					},
				},
			},
			true,
			toState{},
			func(operator *Operator, _ testCase) bool {
				return operator.payload == struct {
					foo string
				} {"bar"}
			},
		},
		"invalid state": {
			invalidState{},
			action{},
			util.StateTransit {},
			false,
			invalidState{},
			nil,
		},
		"invalid action": {
			fromState{},
			invalidAction{},
			util.StateTransit {},
			false,
			fromState{},
			nil,
		},
	}

	for caseName, testCase := range testCases {
		testChan = make(chan string)

		operator := &Operator {
			testCase.fromState,
			testCase.stateTransit,
			dummyWebhookToAction,
			dummyPullRepoNotification,
			dummyGrpcServerRegister,
			nil,
			&sync.Mutex{},
			func() {},
		}
		operator.Dispatch(testCase.action)

		if testCase.toState != nil {
			suite.Assert().Equal(
				testCase.toState,
				operator.state,
				"Case [%s] fails before then func.", caseName,
			)
		}

		if !testCase.testThen {
			continue
		}

		testChan <- "async start"
		_ = <-testChan

		if testCase.afterThen != nil {
			suite.Assert().True(
				testCase.afterThen(operator, testCase),
				"Case [%s] fails after then func.", caseName,
			)
		}
	}
}