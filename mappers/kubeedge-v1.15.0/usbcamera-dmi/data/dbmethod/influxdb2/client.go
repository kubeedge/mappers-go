package influxdb2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kubeedge/usb/pkg/common"
)

type DataBaseConfig struct {
	Influxdb2ClientConfig *Influxdb2ClientConfig `json:"influxdb2ClientConfig,omitempty"`
	Influxdb2DataConfig   *Influxdb2DataConfig   `json:"influxdb2DataConfig,omitempty"`
}

type Influxdb2ClientConfig struct {
	Url    string `json:"url,omitempty"`
	Org    string `json:"org,omitempty"`
	Bucket string `json:"bucket,omitempty"`
}

type Influxdb2DataConfig struct {
	Measurement string            `json:"measurement,omitempty"`
	Tag         map[string]string `json:"tag,omitempty"`
	FieldKey    string            `json:"fieldKey,omitempty"`
}

func NewDataBaseClient(clientConfig json.RawMessage, dataConfig json.RawMessage) (*DataBaseConfig, error) {
	// parse influx database config data
	influxdb2ClientConfig := new(Influxdb2ClientConfig)
	influxdb2DataConfig := new(Influxdb2DataConfig)
	err := json.Unmarshal(clientConfig, influxdb2ClientConfig)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataConfig, influxdb2DataConfig)
	if err != nil {
		return nil, err
	}
	return &DataBaseConfig{
		Influxdb2ClientConfig: influxdb2ClientConfig,
		Influxdb2DataConfig:   influxdb2DataConfig,
	}, nil
}

func (d *DataBaseConfig) InitDbClient() influxdb2.Client {
	var usrtoken string
	usrtoken = os.Getenv("TOKEN")
	client := influxdb2.NewClient(d.Influxdb2ClientConfig.Url, usrtoken)

	return client
}

func (d *DataBaseConfig) CloseSession(client influxdb2.Client) {
	client.Close()
}

func (d *DataBaseConfig) AddData(data *common.DataModel, client influxdb2.Client) error {
	// write device data to influx database
	orgName := d.Influxdb2ClientConfig.Org
	bucketName := d.Influxdb2ClientConfig.Bucket
	ctx := context.Background()
	errMark := "not found"
	org, err := client.OrganizationsAPI().FindOrganizationByName(ctx, orgName)
	if err != nil {
		if strings.Contains(err.Error(), errMark) {
			org, err := client.OrganizationsAPI().CreateOrganizationWithName(ctx, orgName)
			if err != nil {
				return fmt.Errorf("create organization '%s' faild with err:%v", orgName, err)
			}
			_, err = client.BucketsAPI().CreateBucketWithName(ctx, org, bucketName)
			if err != nil {
				return fmt.Errorf("create bucket '%s' faild with err:%v", bucketName, err)
			}
		} else {
			return fmt.Errorf("find organization '%s' faild with err:%v", orgName, err)
		}
	} else {
		buckets, err := client.BucketsAPI().FindBucketsByOrgName(ctx, orgName)
		if err != nil {
			return fmt.Errorf("find buckets faild with err:%v", err)
		}
		var flag = false
		for _, bucket := range *buckets {
			if bucket.Name == bucketName {
				flag = true
				break
			}
		}
		if !flag {
			_, err = client.BucketsAPI().CreateBucketWithName(ctx, org, bucketName)
			if err != nil {
				return fmt.Errorf("create bucket '%s' faild with err:%v", bucketName, err)
			}
		}
	}
	writeAPI := client.WriteAPIBlocking(d.Influxdb2ClientConfig.Org, d.Influxdb2ClientConfig.Bucket)
	p := influxdb2.NewPoint(d.Influxdb2DataConfig.Measurement,
		d.Influxdb2DataConfig.Tag,
		map[string]interface{}{d.Influxdb2DataConfig.FieldKey: data.Value},
		time.Now())
	// write point immediately
	err = writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		klog.V(4).Info("Exit AddData")
		return err
	}
	return nil
}
