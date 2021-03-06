// +build integration

package fileset

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/elastic/beats/libbeat/outputs/elasticsearch"
	"github.com/stretchr/testify/assert"
)

func TestLoadPipeline(t *testing.T) {
	client := elasticsearch.GetTestingElasticsearch()
	client.Request("DELETE", "/_ingest/pipeline/my-pipeline-id", "", nil, nil)

	content := map[string]interface{}{
		"description": "describe pipeline",
		"processors": []map[string]interface{}{
			{
				"set": map[string]interface{}{
					"field": "foo",
					"value": "bar",
				},
			},
		},
	}

	err := loadPipeline(client, "my-pipeline-id", content)
	assert.NoError(t, err)

	status, _, err := client.Request("GET", "/_ingest/pipeline/my-pipeline-id", "", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	// loading again shouldn't actually update the pipeline
	content["description"] = "describe pipeline 2"
	err = loadPipeline(client, "my-pipeline-id", content)
	assert.NoError(t, err)

	status, response, err := client.Request("GET", "/_ingest/pipeline/my-pipeline-id", "", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var res map[string]interface{}
	err = json.Unmarshal(response, &res)
	assert.NoError(t, err)
	assert.Equal(t, "describe pipeline", res["my-pipeline-id"].(map[string]interface{})["description"], string(response))
}

func TestSetupNginx(t *testing.T) {
	client := elasticsearch.GetTestingElasticsearch()
	client.Request("DELETE", "/_ingest/pipeline/filebeat-5.2.0-nginx-access-with_plugins", "", nil, nil)
	client.Request("DELETE", "/_ingest/pipeline/filebeat-5.2.0-nginx-error-pipeline", "", nil, nil)

	modulesPath, err := filepath.Abs("../module")
	assert.NoError(t, err)

	configs := []ModuleConfig{
		{Module: "nginx"},
	}

	reg, err := newModuleRegistry(modulesPath, configs, nil, "5.2.0")
	assert.NoError(t, err)

	err = reg.LoadPipelines(client)
	assert.NoError(t, err)

	status, _, _ := client.Request("GET", "/_ingest/pipeline/filebeat-5.2.0-nginx-access-with_plugins", "", nil, nil)
	assert.Equal(t, 200, status)
	status, _, _ = client.Request("GET", "/_ingest/pipeline/filebeat-5.2.0-nginx-error-pipeline", "", nil, nil)
	assert.Equal(t, 200, status)
}
