package server

import (
	"bufio"
	"context"
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

func livezHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.LivenessCheck(clientCtx, &proto.LivenessCheckRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func readyzHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.ReadinessCheck(clientCtx, &proto.ReadinessCheckRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func metricsHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Metrics(clientCtx, &proto.MetricsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "text/plain; version=0.0.4", grpcResp.Metrics)
}

func clusterHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Cluster(clientCtx, &proto.ClusterRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func createIndexHandlerFunc(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.CreateIndexRequest{}
	if err := marshaler.Unmarshal(body, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.IndexName = c.Param("index_name")

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.CreateIndex(clientCtx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func deleteIndexHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.DeleteIndexRequest{}
	req.IndexName = c.Param("index_name")

	grpcResp, err := client.DeleteIndex(clientCtx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func addDocumentsHandlerFunc(c *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func deleteDocumentsHandlerFunc(c *gin.Context) {
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.DeleteDocuments(clientCtx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}

func searchHandlerFunc(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.SearchRequest{}
	if err := marshaler.Unmarshal(body, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Override with the index name specified by the URI.
	req.IndexName = c.Param("index_name")

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Search(clientCtx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", respBytes)
}
