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

func setMarshaler(marshaler *Marshaler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("marshaler", marshaler)
		c.Next()
	}
}

func getMarshaler(c *gin.Context) (*Marshaler, error) {
	marshalerIntr, ok := c.Get("marshaler")
	if !ok {
		return nil, fmt.Errorf("marshaler does not exist")
	}
	marshaler, ok := marshalerIntr.(*Marshaler)
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

	grpcResp, err := client.LivenessCheck(clientCtx, &proto.LivenessCheckRequest{})
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

	grpcResp, err := client.ReadinessCheck(clientCtx, &proto.ReadinessCheckRequest{})
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

	grpcResp, err := client.Metrics(clientCtx, &proto.MetricsRequest{})
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "text/plain; version=0.0.4", grpcResp.Metrics)
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

	grpcResp, err := client.Cluster(clientCtx, &proto.ClusterRequest{})
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

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func createIndex(c *gin.Context) {
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
		// index_uri is required.
		resp := gin.H{"error": "index_uri is required or unexpected data"}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	req.IndexUri = indexUri

	lockUri, ok := reqMap["lock_uri"].(string)
	if ok {
		// lock_uri is optional
		req.LockUri = lockUri
	}

	indexMapping, ok := reqMap["index_mapping"].(map[string]interface{})
	if ok {
		// Serialize the index_mapping again and set it in the request.
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
		// If num_shards omitted, the number of shards is set to 1.
		numShards = 1
	}
	req.NumShards = uint32(numShards)

	defaultSearchField, ok := reqMap["default_search_field"].(string)
	if ok {
		// default_search_field is optional
		req.DefaultSearchField = defaultSearchField
	}

	defaultAnalyzer, ok := reqMap["default_analyzer"].(map[string]interface{})
	if ok {
		// default_analyzer is optional
		defaultAnalyzerBytes, err := json.Marshal(defaultAnalyzer)
		if err != nil {
			resp := gin.H{"error": "index_uri is not specified or is not a string"}
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		req.DefaultAnalyzer = defaultAnalyzerBytes
	}

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.CreateIndex(clientCtx, req)
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

	grpcResp, err := client.DeleteIndex(clientCtx, &proto.DeleteIndexRequest{
		IndexName: c.Param("index_name"),
	})
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

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func addDocuments(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.AddDocumentsRequest{}
	req.IndexName = c.Param("index_name")
	req.Documents = make([][]byte, 0)

	reader := bufio.NewReader(c.Request.Body)
	for {
		finishReading := false
		// Read a line from the request body
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
			req.Documents = append(req.Documents, docBytes)
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

	marshaler, err := getMarshaler(c)
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

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.DeleteDocuments(clientCtx, req)
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

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func search(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	req := &proto.SearchRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Override with the index name specified by the URI.
	req.IndexName = c.Param("index_name")

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	grpcResp, err := client.Search(clientCtx, req)
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

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}
