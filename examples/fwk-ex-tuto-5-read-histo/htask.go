package main

import (
	"reflect"
	"strings"

	"github.com/go-hep/fwk"
)

type testhsvc struct {
	fwk.TaskBase

	hsvc   fwk.HistSvc
	h1d    fwk.H1D
	stream string
}

func (tsk *testhsvc) Configure(ctx fwk.Context) error {
	var err error

	return err
}

func (tsk *testhsvc) StartTask(ctx fwk.Context) error {
	var err error

	svc, err := ctx.Svc("histsvc")
	if err != nil {
		return err
	}

	tsk.hsvc = svc.(fwk.HistSvc)

	if !strings.HasPrefix(tsk.stream, "/") {
		tsk.stream = "/" + tsk.stream
	}
	if strings.HasSuffix(tsk.stream, "/") {
		tsk.stream = tsk.stream[:len(tsk.stream)-1]
	}

	tsk.h1d, err = tsk.hsvc.BookH1D(tsk.stream+"/h1d-"+tsk.Name(), 100, -10, 10)
	if err != nil {
		return err
	}

	return err
}

func (tsk *testhsvc) StopTask(ctx fwk.Context) error {
	var err error

	h := tsk.h1d.Hist
	if h.Entries() != *evtmax {
		return fwk.Errorf("expected %d entries. got=%d", *evtmax, h.Entries())
	}
	mean := h.Mean()
	if mean != 4.5 {
		return fwk.Errorf("expected mean=%v. got=%v", 4.5, mean)
	}

	rms := h.RMS()
	if rms != 2.8722813232690143 {
		return fwk.Errorf("expected RMS=%v. got=%v", 2.8722813232690143, rms)
	}
	msg := ctx.Msg()
	msg.Infof("histo[%s]: entries=%v mean=%v RMS=%v\n",
		tsk.h1d.ID,
		h.Entries(),
		h.Mean(),
		h.RMS(),
	)

	return err
}

func (tsk *testhsvc) Process(ctx fwk.Context) error {
	var err error
	return err
}

func newtesthsvc(typ, name string, mgr fwk.App) (fwk.Component, error) {
	var err error

	tsk := &testhsvc{
		TaskBase: fwk.NewTask(typ, name, mgr),
		stream:   "",
	}

	err = tsk.DeclProp("Stream", &tsk.stream)
	if err != nil {
		return nil, err
	}

	return tsk, err
}

func init() {
	fwk.Register(reflect.TypeOf(testhsvc{}), newtesthsvc)
}
