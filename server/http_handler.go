package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/marshaler"
	"github.com/mosuka/phalanx/proto"
)

func setClient(client *clients.GRPCIndexClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("client", client)
		c.Next()
	}
}

func getClient(c *gin.Context) (*clients.GRPCIndexClient, error) {
	clientIntr, ok := c.Get("client")
	if !ok {
		return nil, fmt.Errorf("client does not exist")
	}
	client, ok := clientIntr.(*clients.GRPCIndexClient)
	if !ok {
		return nil, fmt.Errorf("client is not an IndexClient")
	}

	return client, nil
}

func setMarshaler(marshaler *marshaler.Marshaler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("marshaler", marshaler)
		c.Next()
	}
}

func getMarshaler(c *gin.Context) (*marshaler.Marshaler, error) {
	marshalerIntr, ok := c.Get("marshaler")
	if !ok {
		return nil, fmt.Errorf("marshaler does not exist")
	}
	marshaler, ok := marshalerIntr.(*marshaler.Marshaler)
	if !ok {
		return nil, fmt.Errorf("marshaler is not a Marshaler")
	}

	return marshaler, nil
}

func livez(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.LivenessCheck(clientCtx, &proto.LivenessCheckRequest{})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func readyz(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.ReadinessCheck(clientCtx, &proto.ReadinessCheckRequest{})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func metrics(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.Metrics(clientCtx, &proto.MetricsRequest{})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "text/plain; version=0.0.4", respBytes)
}

func cluster(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.Cluster(clientCtx, &proto.ClusterRequest{})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func putIndex(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(body, &reqMap); err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.CreateIndexRequest{}

	req.IndexName = c.Param("index_name")

	indexUri, ok := reqMap["index_uri"].(string)
	if !ok {
		resp := gin.H{"error": "index_uri is required or unexpected data"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.IndexUri = indexUri

	lockUri, ok := reqMap["lock_uri"].(string)
	if !ok {
		resp := gin.H{"error": "lock_uri is required or unexpected data"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.LockUri = lockUri

	indexMapping, ok := reqMap["index_mapping"].(map[string]interface{})
	if ok {
		indexMappingBytes, err := json.Marshal(indexMapping)
		if err != nil {
			resp := gin.H{"error": "index_uri is not specified or is not a string"}
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		req.IndexMapping = indexMappingBytes
	}

	numShards, ok := reqMap["num_shards"].(float64)
	if !ok {
		resp := gin.H{"error": "num_shards is required or unexpected data"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.NumShards = uint32(numShards)

	grpcResp, err := client.CreateIndex(clientCtx, req)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func deleteIndex(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.DeleteIndex(clientCtx, &proto.DeleteIndexRequest{
		IndexName: c.Param("index_name"),
	})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func putDocuments(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.AddDocumentsRequest{}
	req.IndexName = c.Param("index_name")
	req.Documents = make([]*proto.Document, 0)

	reader := bufio.NewReader(c.Request.Body)
	for {
		finishReading := false
		docBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				finishReading = true
			} else {
				resp := gin.H{"error": err.Error()}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
		}
		if len(docBytes) > 0 {
			docMap := make(map[string]interface{})
			if err := json.Unmarshal(docBytes, &docMap); err != nil {
				resp := gin.H{"error": err.Error()}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
			id, ok := docMap["id"].(string)
			if !ok {
				resp := gin.H{"error": "document id does not exist or is not a string"}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
			fields := docMap["fields"].(map[string]interface{})
			if !ok {
				resp := gin.H{"error": fmt.Sprintf("id: %s fields do not exist or is not a map[string]interface{}", id)}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
			fieldsBytes, err := json.Marshal(fields)
			if err != nil {
				resp := gin.H{"error": err.Error()}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}

			doc := &proto.Document{
				Id:     id,
				Fields: fieldsBytes,
			}
			req.Documents = append(req.Documents, doc)
		}
		if finishReading {
			break
		}
	}

	grpcResp, err := client.AddDocuments(clientCtx, req)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func deleteDocuments(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.DeleteDocumentsRequest{}
	req.IndexName = c.Param("index_name")
	req.Ids = make([]string, 0)

	reader := bufio.NewReader(c.Request.Body)
	for {
		finishReading := false
		docIdBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				finishReading = true
			} else {
				resp := gin.H{"error": err.Error()}
				c.JSON(http.StatusInternalServerError, resp)
				return
			}
		}
		if len(docIdBytes) > 0 {
			req.Ids = append(req.Ids, strings.TrimSpace(string(docIdBytes)))
		}
		if finishReading {
			break
		}
	}

	grpcResp, err := client.DeleteDocuments(clientCtx, req)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func search(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(body, &reqMap); err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.SearchRequest{}
	req.IndexName = c.Param("index_name")

	query, ok := reqMap["query"].(string)
	if !ok {
		resp := gin.H{"error": "query does not exist or is not a string"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.Query = query

	field, ok := reqMap["field"].(string)
	if !ok {
		resp := gin.H{"error": "field does not exist or is not a string"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.Field = field

	boost, ok := reqMap["boost"].(float64)
	if !ok {
		resp := gin.H{"error": "boost does not exist or is not a number"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.Boost = boost

	start, ok := reqMap["start"].(float64)
	if !ok {
		resp := gin.H{"error": "start does not exist or is not a number"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.Start = int32(start)

	num, ok := reqMap["num"].(float64)
	if !ok {
		resp := gin.H{"error": "start does not exist or is not a number"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.Num = int32(num)

	grpcResp, err := client.Search(clientCtx, req)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}
