//
// November 2016, cisco
//
// Copyright (c) 2016 by cisco Systems, Inc.
// All rights reserved.
//
//
// Input node used to replay streaming telemetry archives. Archives
// can be recoded using 'tap' output module with "raw = true" set.
//
// Tests for replay module are in tap_test.go
//
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	math "math"
	"os"
	"time"

	telem "github.com/cisco/bigmuddy-network-telemetry-proto/proto_go"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	REPLAY_DELAY_DEFAULT_USEC = 200000
)

// realTimeState keeps track of the timing when simulateRealTime option is enabled.
type realTimeState struct {
	// The first timestamp seen in the sample
	originalStartTime time.Time

	// intervalTime is the interval duration between the first and second sample.
	// it is required to calculate each following iteration when we loop over the initial data sample
	// and it is only set once
	intervalTime *time.Duration

	// The time at which this test was started (i.e "now")
	testStartTime time.Time

	// The latest time that was calculated. This is refreshed everytime
	// "computeNextTime" is called
	latestTimeSet time.Time

	// The message offset to be applied. This get refreshed on every loop
	offset time.Duration
}

func newRealTimeState(originalStartTime time.Time, loopBacklog time.Duration) *realTimeState {
	now := time.Now().Add(-1 * loopBacklog)
	return &realTimeState{
		originalStartTime: originalStartTime,
		testStartTime:     now,
		offset:            now.Sub(originalStartTime),
	}
}

func newRealTimeStateMsec(originalStartTimeMsec uint64, loopBacklog time.Duration) *realTimeState {
	return newRealTimeState(uint64time(originalStartTimeMsec), loopBacklog)
}

func (p *realTimeState) String() string {
	intervalTimeStr := "na"
	if nil != p.intervalTime {
		intervalTimeStr = p.intervalTime.String()
	}
	return fmt.Sprintf("first sample @[%.28s], started @[%.28s], offset: [%s], latestTimeSet: [%s], intervalTime: [%s]", p.originalStartTime.String(), p.testStartTime.String(), p.offset.String(), p.latestTimeSet.String(), intervalTimeStr)
}

func (p *realTimeState) computeNextTime(oTime time.Time) time.Time {
	if nil == p.intervalTime && oTime != p.originalStartTime {
		interval := oTime.Sub(p.originalStartTime)
		p.intervalTime = &interval
	}
	p.latestTimeSet = oTime.Add(p.offset)
	return p.latestTimeSet
}

func (p *realTimeState) computeNextTimeMsec(oTimeMsec uint64) uint64 {
	nextTime := p.computeNextTime(uint64time(oTimeMsec))
	return uint64(nextTime.UnixNano() / 1000000)
}

// nextLoop is called whenever an input file has reached the end and is about to be looped.
// We update the offset for the next iteration
func (p *realTimeState) nextLoop() {
	p.offset = p.latestTimeSet.Sub(p.originalStartTime)
	if nil != p.intervalTime {
		p.offset += *p.intervalTime
	}
}

//
// Module implementing inputNodeModule interface allowing REPLAY to read
// binary dump for replay.
type replayInputModule struct {
	name             string
	logctx           *log.Entry
	filename         string
	logData          bool
	firstN           int
	loop             bool
	delayUsec        int
	done             bool
	count            int
	simulateRealTime bool
	loopBacklog      time.Duration
	ctrlChan         chan *ctrlMsg
	dataChans        []chan<- dataMsg
}

func replayInputModuleNew() inputNodeModule {
	return &replayInputModule{}
}

func (t *replayInputModule) String() string {
	return fmt.Sprintf("%s:%s", t.name, t.filename)
}

func (t *replayInputModule) replayInputFeederLoop() error {

	var stats msgStats
	var tickIn time.Duration
	var tick <-chan time.Time

	err, parser := getNewEncapParser(t.name, "st", t)
	if err != nil {
		t.logctx.WithError(err).Error(
			"Failed to open get parser, STOP")
		t.done = true
	}

	err, gpbCodec := getCodec("replayer_gdb_codec", ENCODING_GPB_KV)
	if err != nil {
		t.logctx.WithError(err).Error("Failed to get GPB_KV codec")
		t.done = true
	}

	var loopRTState *realTimeState

	f, err := os.Open(t.filename)
	if err != nil {
		t.logctx.WithError(err).Error(
			"Failed to open file with binary dump of telemetry. " +
				"Dump should be produced with 'tap' output, 'raw=true', STOP")
		t.done = true
	} else {
		defer f.Close()

		if t.delayUsec != 0 {
			tickIn = time.Duration(t.delayUsec) * time.Microsecond
		} else {
			//
			// We still tick to get control channel a look in.
			tickIn = time.Nanosecond
		}
		tick = time.Tick(tickIn)
	}

	for {

		select {

		case <-tick:

			if t.done {
				// Waiting for exit, we're done here.
				continue
			}

			// iterate until a message is produced (header, payload)
			var i int
			for {
				i = i + 1
				err, buffer := parser.nextBlockBuffer()
				if err != nil {
					t.logctx.WithError(err).WithFields(
						log.Fields{
							"iteration": i,
							"nth_msg":   t.count,
						}).Error("Failed to fetch buffer, STOP")
					t.done = true
					return err
				}

				readn, err := io.ReadFull(f, *buffer)
				if err != nil {
					if err != io.EOF {
						t.logctx.WithError(err).WithFields(
							log.Fields{
								"iteration": i,
								"nth_msg":   t.count,
							}).Error("Failed to read next buffer, STOP")
						t.done = true
						return err
					}

					if !t.loop {
						//
						// We're done.
						return nil
					}

					t.logctx.Debug("restarting from start of message archive")
					_, err := f.Seek(0, 0)
					if err != nil {
						t.logctx.WithError(err).WithFields(
							log.Fields{
								"iteration": i,
								"nth_msg":   t.count,
							}).Error("Failed to go back to start, STOP")
						t.done = true
						return err
					}

					if nil != loopRTState {
						loopRTState.nextLoop()
					}

					//
					// Because we're starting from scratch, and we
					// need to get a new parser to restart parser
					// state machine along with data.
					err, parser = getNewEncapParser(t.name, "st", t)
					continue
				}
				err, msgs := parser.nextBlock(*buffer, nil)
				if err != nil {
					t.logctx.WithError(err).WithFields(
						log.Fields{
							"iteration": i,
							"nth_msg":   t.count,
							"read_in":   readn,
							"len":       len(*buffer),
							"msg":       hex.Dump(*buffer),
						}).Error(
						"Failed to decode next block, STOP")
					t.done = true
					return err
				}

				if t.logData {
					t.logctx.WithFields(log.Fields{
						"iteration":    i,
						"nth_msg":      t.count,
						"dataMsgCount": len(msgs),
						"len":          len(*buffer),
						"msg":          hex.Dump(*buffer),
					}).Debug("REPLAY input logdata")
				}

				if msgs == nil {
					//
					// We probably just read a header
					continue
				}

				// When we get here, we have complete messages to process
				for _, msg := range msgs {

					// for GPB messages, we support updating the timestamp (more format to come later)
					if m, ok := msg.(*dataMsgGPB); ok {

						if t.simulateRealTime {

							if m.cachedDecode == nil {
								t.logctx.Errorf("Failed to decode GPB for replay: Content table is nil?")
								continue
							}

							if loopRTState == nil {
								loopRTState = newRealTimeStateMsec(m.cachedDecode.GetMsgTimestamp(), t.loopBacklog)
							}

							// The the next timestmap to use
							nextTimestamp := loopRTState.computeNextTimeMsec(m.cachedDecode.GetMsgTimestamp())

							timeUntilNow := uint64time(nextTimestamp).Sub(time.Now())
							if timeUntilNow > 0 {
								t.logctx.Infof("Replay loop as reached current time. Halting replay loop for %s before generating next sample", timeUntilNow.String())
								time.Sleep(timeUntilNow)
							}
							t.logctx.Infof("UINT64 start time %s, nextTime: %s",
								uint64time(m.cachedDecode.GetMsgTimestamp()).String(),
								uint64time(nextTimestamp).String())

							target := &telem.Telemetry{}
							target.NodeId = m.cachedDecode.NodeId
							target.Subscription = m.cachedDecode.Subscription
							target.CollectionStartTime = m.cachedDecode.CollectionStartTime
							target.MsgTimestamp = nextTimestamp
							target.CollectionEndTime = m.cachedDecode.CollectionEndTime
							target.EncodingPath = m.cachedDecode.EncodingPath

							for _, kv := range m.cachedDecode.DataGpbkv {
								newKv := &telem.TelemetryField{
									Timestamp:   loopRTState.computeNextTimeMsec(kv.GetTimestamp()),
									Name:        kv.Name,
									ValueByType: kv.ValueByType,
									Fields:      kv.Fields,
								}

								target.DataGpbkv = append(target.DataGpbkv, newKv)
							}
							newContent, err := proto.Marshal(target)
							if err != nil {
								t.logctx.WithError(err).Error("Failed to marshal proto message")
							}
							err, dataMsg := codecGPBBlockAndTelemMsgToDataMsgs(gpbCodec.(*codecGPB), t, newContent, target)
							if err != nil {
								t.logctx.WithError(err).Error("Failed to encode message with updated timestamp")
							}

							if len(dataMsg) > 1 {
								t.logctx.Errorf("While re-encoding message with new timestamp, found %d msg to encode while only one was expected", len(dataMsg))

							}
							msg = dataMsg[0]

						}

					}

					for _, dataChan := range t.dataChans {
						if len(dataChan) == cap(dataChan) {
							t.logctx.Error("Input overrun (replace with counter)")
							continue
						}
						dataChan <- msg
					}

				}

				t.count = t.count + 1
				if t.count == t.firstN && t.firstN != 0 {
					t.logctx.Debug("dumped all messages expected")
					t.done = true
				}

				break
			}

		case msg := <-t.ctrlChan:
			switch msg.id {
			case REPORT:
				t.logctx.Debug("report request")
				content, _ := json.Marshal(stats)
				resp := &ctrlMsg{
					id:       ACK,
					content:  content,
					respChan: nil,
				}
				msg.respChan <- resp

			case SHUTDOWN:
				t.logctx.Info("REPLAY input loop, rxed SHUTDOWN, shutting down")

				resp := &ctrlMsg{
					id:       ACK,
					respChan: nil,
				}
				msg.respChan <- resp

				return nil

			default:
				t.logctx.Error("REPLAY input loop, unknown ctrl message")
			}
		}
	}
}

func (t *replayInputModule) replayInputFeederLoopSticky() {
	t.logctx.Debug("Starting REPLAY feeder loop")
	for {
		err := t.replayInputFeederLoop()
		if err == nil {
			t.logctx.Debug("REPLAY feeder loop done, exit")
			break
		} else {
			// retry
			time.Sleep(time.Second)
			if t.done {
				t.logctx.WithFields(log.Fields{
					"nth_msg": t.count,
				}).Debug("idle, waiting for shutdown")
			} else {
				t.logctx.Debug("Restarting REPLAY feeder loop")
			}
		}
	}
}

func uint64time(ts uint64) time.Time {

	tsSec := int64(ts / 1000)
	tsNsec := int64(math.Mod(float64(ts), 1000) * float64(time.Microsecond.Nanoseconds()))

	return time.Unix(tsSec, tsNsec)

}
func (t *replayInputModule) configure(
	name string,
	nc nodeConfig,
	dataChans []chan<- dataMsg) (error, chan<- *ctrlMsg) {

	var err error

	t.filename, err = nc.config.GetString(name, "file")
	if err != nil {
		return err, nil
	}

	t.name = name
	t.logData, _ = nc.config.GetBool(name, "logdata")
	t.firstN, _ = nc.config.GetInt(name, "firstn")
	if t.firstN == 0 {
		var err error
		t.loop, _ = nc.config.GetBool(name, "loop")
		t.simulateRealTime, _ = nc.config.GetBool(name, "simulateRealTime")
		loopStartBacklogStr, _ := nc.config.GetString(name, "loopBacklog")

		t.loopBacklog, err = time.ParseDuration(loopStartBacklogStr)
		if err != nil {
			log.Errorf("%s is an invalid format for a time.Duration: %s", loopStartBacklogStr, err.Error())
			panic(err)
		}
	}

	t.delayUsec, err = nc.config.GetInt(name, "delayusec")
	if err != nil {
		//
		// Default to a sensible-ish value for replay to
		// avoid overwhelming output stages. Note that
		// we can still overwhelm output stages if we
		// want to do so explicitly i.e. set to zero.
		t.delayUsec = REPLAY_DELAY_DEFAULT_USEC
	}

	t.ctrlChan = make(chan *ctrlMsg)
	t.dataChans = dataChans

	t.logctx = logger.WithFields(log.Fields{
		"name":      t.name,
		"file":      t.filename,
		"logdata":   t.logData,
		"firstN":    t.firstN,
		"delayUsec": t.delayUsec,
		"loop":      t.loop,
	})

	go t.replayInputFeederLoopSticky()

	return nil, t.ctrlChan
}
