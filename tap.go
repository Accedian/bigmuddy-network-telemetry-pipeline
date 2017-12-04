//
// February 2016, cisco
//
// Copyright (c) 2016 by cisco Systems, Inc.
// All rights reserved.
//
//
// Output node used to tap pipeline for troubleshooting
//
package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//
// Module implementing outputNodeModule interface.
type tapOutputModule struct {
	name             string
	filename         string
	countOnly        bool
	rawDump          bool
	printSummary     bool
	pm_file_template string
	forceTextOutput  bool
	streamSpec       *dataMsgStreamSpec
	dataChannelDepth int
	ctrlChan         chan *ctrlMsg
	dataChan         chan dataMsg
}

func tapOutputModuleNew() outputNodeModule {
	return &tapOutputModule{
		printSummary: true,
	}
}

func (t *tapOutputModule) SplitPMFiles() bool {
	return t.pm_file_template != ""
}

func (t *tapOutputModule) tapOutputFeederLoop() {
	var stats msgStats
	var hexOnly bool

	// Period, in seconds, to dump stats if only counting.
	const TIMEOUT = 10
	timeout := make(chan bool, 1)

	if !t.streamSpec.dataMsgStreamSpecTextBased() && !t.forceTextOutput {
		hexOnly = true
	}

	logctx := logger.WithFields(
		log.Fields{
			"name":       t.name,
			"filename":   t.filename,
			"countonly":  t.countOnly,
			"streamSpec": t.streamSpec,
		})

	// Prepare dump file for writing
	f, err := os.OpenFile(t.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0660)
	if err != nil {
		logctx.WithError(err).Error("Tap failed to open dump file")
		return
	}
	defer f.Close()

	outWriter := bufio.NewWriter(f)
	defer outWriter.Flush()

	logctx.Info("Starting up tap")

	if t.countOnly {
		go func() {
			time.Sleep(TIMEOUT * time.Second)
			timeout <- true
		}()
	}

	for {

		func() {

			w := outWriter

			select {

			case <-timeout:

				go func() {
					time.Sleep(TIMEOUT * time.Second)
					timeout <- true
				}()
				w.WriteString(fmt.Sprintf(
					"%s:%s: rxed msgs: %v\n",
					t.name, time.Now().Local(), stats.MsgsOK))
				w.Flush()

			case msg, ok := <-t.dataChan:

				if !ok {
					// Channel has been closed. Our demise
					// is near. SHUTDOWN is likely to be
					t.dataChan = nil
					break
				}

				if t.countOnly {
					stats.MsgsOK++
					break
				}

				dM := msg
				description := dM.getDataMsgDescription()

				if t.rawDump {
					err, b := dM.produceByteStream(dataMsgStreamSpecDefault)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Tap failed to produce raw message")
						break
					}
					err, enc := dataMsgStreamTypeToEncoding(
						dM.getDataMsgStreamType())
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Tap failed to identify encoding")
						break
					}
					err, encst := encapSTFromEncoding(enc)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Tap failed to identify encap st")
						break
					}

					//
					// We should really push this into co side of codec.
					hdr := encapSTHdr{
						MsgType:       ENC_ST_HDR_MSG_TYPE_TELEMETRY_DATA,
						MsgEncap:      encst,
						MsgHdrVersion: ENC_ST_HDR_VERSION,
						Msgflag:       ENC_ST_HDR_MSG_FLAGS_NONE,
						Msglen:        uint32(len(b)),
					}
					err = binary.Write(w, binary.BigEndian, hdr)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Errorf("Tap failed to write binary hdr %+v", hdr)
						break
					}

					_, err = w.Write(b)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Tap failed to write binary message")
						break
					}

					break
				}

				// OK. We're ready to dump something largely human readable.
				errStreamType, b := dM.produceByteStream(t.streamSpec)
				if errStreamType != nil {
					err, b = dM.produceByteStream(dataMsgStreamSpecDefault)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Tap failed to dump message")
						stats.MsgsNOK++
						break
					}
				} else if b == nil {
					break
				}
				stats.MsgsOK++

				// If we must generate a new PM file per data message, lets create the file there
				if t.SplitPMFiles() {
					pmFileName, err := msg.generatePMFileName(t.pm_file_template)
					if err != nil {
						logctx.WithError(err).WithFields(
							log.Fields{
								"msg": description,
							}).Error("Failed to generate PM file name")
						break
					}
					if conductor.Debug {
						fmt.Printf("Creating PM file %s\n", pmFileName)
					}

					// Prepare dump file for writing
					pmFile, err := os.OpenFile(pmFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND,
						0660)

					if err != nil {
						logctx.WithError(err).Error("Tap failed to open dump file")
						return
					}

					defer func() {
						pmFile.Close()
					}()

					w = bufio.NewWriter(pmFile)
				}

				if t.printSummary {
					w.WriteString(fmt.Sprintf(
						"\n------- %v -------\n", time.Now()))
					w.WriteString(fmt.Sprintf("Summary: %s\n", description))
				}
				if hexOnly || errStreamType != nil {
					if errStreamType != nil {
						w.WriteString(fmt.Sprintf(
							"Requested stream type failed: [%v]\n", errStreamType))
					}
					w.WriteString(hex.Dump(b))
				} else {

					// Everything is clean, lets dump the record
					w.Write(b)
				}
				w.Flush()

			case msg := <-t.ctrlChan:
				switch msg.id {
				case REPORT:
					content, _ := json.Marshal(stats)
					resp := &ctrlMsg{
						id:       ACK,
						content:  content,
						respChan: nil,
					}
					msg.respChan <- resp

				case SHUTDOWN:

					w.Flush()
					logctx.Info("tap feeder loop, rxed SHUTDOWN")

					//
					// Dump detailed stats here

					resp := &ctrlMsg{
						id:       ACK,
						respChan: nil,
					}
					msg.respChan <- resp
					return

				default:
					logctx.Error("tap feeder loop, unknown ctrl message")
				}
			}
		}()
	}
}

//
// Setup a tap output module so we can see what is going on.
func (t *tapOutputModule) configure(name string, nc nodeConfig) (
	error, chan<- dataMsg, chan<- *ctrlMsg) {

	var err error

	t.name = name

	t.filename, err = nc.config.GetString(name, "file")
	if err != nil {
		return err, nil, nil
	}

	t.dataChannelDepth, err = nc.config.GetInt(name, "datachanneldepth")
	if err != nil {
		t.dataChannelDepth = DATACHANNELDEPTH
	}

	// If not set, will default to false, but let's be clear.
	t.countOnly, _ = nc.config.GetBool(name, "countonly")

	// Looking for a raw dump?
	t.rawDump, _ = nc.config.GetBool(name, "raw")
	if t.rawDump {
		if t.countOnly {
			logger.WithError(err).WithFields(
				log.Fields{"name": name}).Error(
				"tap config: 'countonly' is incompatible with 'raw'")
			return err, nil, nil
		}
		_, err = nc.config.GetString(name, "encoding")
		if err == nil {
			logger.WithError(err).WithFields(
				log.Fields{"name": name}).Error(
				"tap config: 'encoding' is incompatible with 'raw'")
			return err, nil, nil
		}
	}
	nc.config.GetString(name, "encoding")

	// Pick output stream type
	err, t.streamSpec = dataMsgStreamSpecFromConfig(nc, name)
	if err != nil {
		logger.WithError(err).WithFields(
			log.Fields{
				"name": name,
			}).Error("'encoding' option for tap output")
		return err, nil, nil
	}

	printSummary, err := nc.config.GetBool(name, "print_summary")
	if err != nil && nc.config.HasOption(name, "print_summary") {
		logger.WithError(err).WithFields(
			log.Fields{
				"name": name,
			}).Error("'print_summary' option for tap output")
		return err, nil, nil
	}
	t.printSummary = printSummary

	pm_file_template, err := nc.config.GetString(name, "pm_file_template")
	if err != nil && nc.config.HasOption(name, "pm_file_template") {
		logger.WithError(err).WithFields(
			log.Fields{
				"name": name,
			}).Error("'pm_file_template' option for tap output")
		return err, nil, nil
	}
	t.pm_file_template = convertStrVarToTemplArgs(pm_file_template)

	forceTextOutput, err := nc.config.GetBool(name, "force_text_output")
	if err != nil && nc.config.HasOption(name, "force_text_output") {
		logger.WithError(err).WithFields(
			log.Fields{
				"name": name,
			}).Error("'force_text_output' option for tap output")
		return err, nil, nil
	}
	t.forceTextOutput = forceTextOutput

	if t.forceTextOutput && t.rawDump {
		logger.WithFields(
			log.Fields{
				"name": name,
			}).Error("Both 'force_text_output' and 'raw' set to true in tap configuration. These options are mutually exclusive")
	}

	// Setup control and data channels
	t.ctrlChan = make(chan *ctrlMsg)
	t.dataChan = make(chan dataMsg, t.dataChannelDepth)

	go t.tapOutputFeederLoop()

	return nil, t.dataChan, t.ctrlChan

}

// A hacky function to replace ${...} in variable names, escaping the templating applied on
// configuration file
func convertStrVarToTemplArgs(varStr string) string {
	vOpen := strings.Replace(varStr, "$(", "{{ .", -1)
	return strings.Replace(vOpen, ")", " }}", -1)
}
