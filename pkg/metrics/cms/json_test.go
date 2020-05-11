package cms

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCMSMetricSource_GetExternalMetric(t *testing.T) {
	resp:=`{\"timestamp\":1588915980000,\"userId\":\"1240538168824185\",\"groupId\":\"37224004\",\"Sum\":0},{\"timestamp\":1588916040000,\"userId\":\"1240538168824185\",\"groupId\":\"37224004\",\"Sum\":0}`
	var res []DataPoint
	json.Unmarshal([]byte(resp),&res)
	fmt.Println(res)
}