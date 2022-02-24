package server

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/errors"
	"github.com/mosuka/phalanx/mapping"
	"github.com/mosuka/phalanx/proto"
)

//go:embed static/*
var staticFS embed.FS

func setClient(client *clients.GRPCIndexClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("client", client)
		c.Next()
	}
}

func getClient(ctx *gin.Context) (*clients.GRPCIndexClient, error) {
	clientIntr, ok := ctx.Get("client")
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

func getMarshaler(ctx *gin.Context) (*Marshaler, error) {
	marshalerIntr, ok := ctx.Get("marshaler")
	if !ok {
		return nil, fmt.Errorf("marshaler does not exist")
	}
	marshaler, ok := marshalerIntr.(*Marshaler)
	if !ok {
		return nil, fmt.Errorf("marshaler is not a Marshaler")
	}

	return marshaler, nil
}

func staticHandlerFunc(ctx *gin.Context) {
	staticServer := http.FileServer(http.FS(staticFS))
	staticServer.ServeHTTP(ctx.Writer, ctx.Request)
}

func livezHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.LivenessCheck(clientCtx, &proto.LivenessCheckRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func readyzHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.ReadinessCheck(clientCtx, &proto.ReadinessCheckRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func metricsHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Metrics(clientCtx, &proto.MetricsRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "text/plain; version=0.0.4", grpcResp.Metrics)
}

func clusterHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Cluster(clientCtx, &proto.ClusterRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func createIndexHandlerFunc(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.CreateIndexRequest{}
	if err := marshaler.Unmarshal(body, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.IndexName = ctx.Param("index_name")

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.CreateIndex(clientCtx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func deleteIndexHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.DeleteIndexRequest{}
	req.IndexName = ctx.Param("index_name")

	grpcResp, err := client.DeleteIndex(clientCtx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func addDocumentsHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.AddDocumentsRequest{}
	req.IndexName = ctx.Param("index_name")
	req.Documents = make([]*proto.Document, 0)

	reader := bufio.NewReader(ctx.Request.Body)
	for {
		finishReading := false
		// Read a line from the request body
		fieldsBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				finishReading = true
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(fieldsBytes) > 0 {
			if strings.Trim(string(fieldsBytes), "\n") == "" {
				// Empty line will be skipped.
				continue
			}

			// Deserialize bytes to fields map.
			fields := make(map[string]interface{})
			if err := json.Unmarshal(fieldsBytes, &fields); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Get document ID
			docID, ok := fields[mapping.IdFieldName].(string)
			if !ok {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDocumentIdDoesNotExist.Error()})
				return

			}

			doc := &proto.Document{
				Id:     docID,
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func deleteDocumentsHandlerFunc(ctx *gin.Context) {
	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	req := &proto.DeleteDocumentsRequest{}
	req.IndexName = ctx.Param("index_name")
	req.Ids = make([]string, 0)

	reader := bufio.NewReader(ctx.Request.Body)
	for {
		finishReading := false
		docIdBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF || err == io.ErrClosedPipe {
				finishReading = true
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(docIdBytes) > 0 {
			if strings.Trim(string(docIdBytes), "\n") == "" {
				// Empty line will be skipped.
				continue
			}

			req.Ids = append(req.Ids, strings.TrimSpace(string(docIdBytes)))
		}
		if finishReading {
			break
		}
	}

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.DeleteDocuments(clientCtx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}

func searchHandlerFunc(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshaler, err := getMarshaler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req := &proto.SearchRequest{}
	if err := marshaler.Unmarshal(body, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Override with the index name specified by the URI.
	req.IndexName = ctx.Param("index_name")

	clientCtx, clientCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer clientCancel()

	client, err := getClient(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grpcResp, err := client.Search(clientCtx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respBytes, err := marshaler.Marshal(grpcResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "application/json", respBytes)
}
