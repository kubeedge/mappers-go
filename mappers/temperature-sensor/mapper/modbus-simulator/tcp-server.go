package modbus_simulator

import (
	"github.com/thinkgos/gomodbus"
	"k8s.io/klog/v2"
)

func InitModbusSimulator(address string) {
	srv := modbus.NewTCPServer()
	// srv.LogMode(true)
	srv.AddNodes(
		modbus.NewNodeRegister(1, 0, 1, 0, 0, 0,0,0,1),
	)
	defer srv.Close()
	if err := srv.ListenAndServe(address); err != nil {
		panic(err)
	}
	klog.Infoln("Modbus simulator started, listening on :", address)
}
