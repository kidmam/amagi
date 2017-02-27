package elasticsearch

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/b-eee/amagi/services/database"
	"gopkg.in/mgo.v2/bson"

	utils "github.com/b-eee/amagi"
	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	// ESConn main elasticsearch connection var
	ESConn *elastic.Client

	// DatastoreNode datastore node labels
	DatastoreNode = "ds:Datastore"
	// ProjectNode project node name
	ProjectNode = "pj:Project"
	// WorkspaceNode workspace nodename
	WorkspaceNode = "ws:Workspace"
	// DatastoreFields datastore fields
	DatastoreFields = "fs:Fields"
	// FieldNodeLabel field node label
	FieldNodeLabel = "f:Field"
	// RoleNodeLabel role node label name
	RoleNodeLabel = "r:Role"
	// UserNodeLabel user node name label
	UserNodeLabel = "u:User"

	// DatastoreCollection collection name
	DatastoreCollection = "data_stores"
)

type (
	// ESSearchReq elasticsearch search request
	ESSearchReq struct {
		IndexName  string
		Type       string
		Context    context.Context
		BodyJSON   interface{}
		FileBase64 string

		SearchName   string
		SearchField  string
		SearchValues interface{}

		SortField string
		SortAsc   bool

		UserID        string
		UserBasicInfo UserBasicInfo

		UpdateChanges map[string]interface{}
	}

	// UserBasicInfo user basic info
	UserBasicInfo struct {
		ID         string
		AccessKeys []string
	}

	// Testing test struct
	Testing struct {
		User    string
		Message string
	}

	// DistinctItems distincted item values for elasticsearch
	DistinctItems struct {
		ID           string       `bson:"_id" json:"_id"`
		DistinctItem DistinctItem `bson:"distinct_item" json:"distinct_item"`
	}

	// DistinctItem unwinded item
	DistinctItem struct {
		WID   string `bson:"w_id" json:"w_id"`
		PID   string `bson:"p_id" json:"p_id"`
		DID   string `bson:"d_id" json:"d_id"`
		QID   string `bson:"q_id" json:"q_id"`
		IID   string `bson:"i_id" json:"i_id"`
		FID   string `bson:"f_id" json:"f_id"`
		AID   string `bson:"a_id" json:"a_id"`
		HID   string `bson:"h_id" json:"h_id"`
		Index string `json:"index"`
		Value string `bson:"value" json:"value"`

		Attachment struct {
			Content interface{} `json:"content,omitempty"`
		} `json:"attachment,omitempty"`
	}

)

// ESCreateIndex elasticsearch create index
func ESCreateIndex(indexName string) error {
	ctx := context.Background()
	// Create an index
	if _, err := database.ESGetConn().CreateIndex(indexName).Do(ctx); err != nil {
		// Handle error
		panic(err)
	}

	return nil
}

// ESAddDocument add document to the index
func (req *ESSearchReq) ESAddDocument() error {
	// indexname should be lowercase
	indexName := strings.ToLower(req.IndexName)

	if exists, err := database.ESGetConn().IndexExists(indexName).Do(CreateContext()); !exists || err != nil {
		utils.Error(fmt.Sprintf("index does not exists! err=%v creating index.. %v", err, indexName))
		if err := ESCreateIndex(strings.ToLower(indexName)); err != nil {
			return err
		}
	}

	s := time.Now()
	if _, err := database.ESGetConn().Index().
		Index(indexName).
		Type(req.Type).
		BodyJson(req.BodyJSON).
		Refresh("true").
		Do(CreateContext()); err != nil {

		utils.Error(fmt.Sprintf("error ESAddDocument %v", err))
		return err
	}

	utils.Info(fmt.Sprintf("ESaddDocument took: %v index=%v", time.Since(s), indexName))
	return nil
}

// ESDeleteDocument delete document
func (req *ESSearchReq) ESDeleteDocument() error {
	del := elastic.NewMatchQuery("i_id", req.BodyJSON.(DistinctItem).IID)

	res, err := elastic.NewDeleteByQueryService(database.ESGetConn()).
		// for multiple index search query, pass in slice of string
		Index(strings.Split(req.IndexName, ",")...).
		Query(del).
		Do(CreateContext())
	if err != nil {
		utils.Error(fmt.Sprintf("error ESDeleteDocument %v", err))
		return err
	}

	fmt.Println("deleted: ", res.Deleted)
	if res.Deleted == 0 {
		return fmt.Errorf("deleted: %v", res.Deleted)
	}

	return nil
}

// ESUpdateDocument update elasticSearch document
func (req *ESSearchReq) ESUpdateDocument() error {

	// TODO handle in switch instead
	// req.ESHTTPItemUpdate()
	return nil

}

type concurrentSearch struct {
	Query elastic.Query
	Field string
}

// ESTermQuery new term query
// manual settings for setting default tokenizer for kuromoji
// $ curl -u elastic -XPOST 'http://104.198.115.53:9400/datastore/_close'
// $ curl -u elastic -XPUT 'http://104.198.115.53:9400/datastore/_settings?preserve_existing=true' -d '{   "index.analysis.analyzer.default.tokenizer" : "kuromoji",   "index.analysis.analyzer.default.type" : "custom" }'
// $ curl -u elastic -XPOST 'http://104.198.115.53:9400/datastore/_open'
func (req *ESSearchReq) ESTermQuery(result *elastic.SearchResult) (*elastic.SearchResult, error) {
	// joinedText := buildRegexpString(req.SearchValues)
	// query := elastic.NewRegexpQuery(req.SearchField, joinedText).
	// 	Boost(1.2).Analyzer("analyzer")

	query := elastic.NewSimpleQueryStringQuery(fmt.Sprintf("%v", req.SearchValues)).Field("value").Field("attachment.content").AnalyzeWildcard(true).Analyzer("kuromoji")

	searchResult, err := database.ESGetConn().Search().
		Index("_all").
		Highlight(ResultHighlighter(req.SearchField)).
		Query(query).
		From(0).
		Size(50).
		Do(CreateContext())
	if err != nil {
		return nil, err
	}

	utils.Info(fmt.Sprintf("ESTermQuery took: %v ms hits: %v", searchResult.TookInMillis, searchResult.Hits.TotalHits))
	return searchResult, nil
}

// ResultHighlighter create result highlighter
func ResultHighlighter(field string) *elastic.Highlight {
	return elastic.NewHighlight().
		Fields(elastic.NewHighlighterField(field)).
		PreTags("<em class='searched_em'>").PostTags("</em>")
}

// ESBulkDeleteDocuments bulk delete elasticsearch document
func ESBulkDeleteDocuments(requests ...ESSearchReq) error {
	for _, req := range requests {
		if err := req.ESDeleteDocument(); err != nil {
			continue
		}
	}

	return nil
}

// ESBulkAddDocuments bulk delete elasticsearch document
func ESBulkAddDocuments(requests ...ESSearchReq) error {
	for _, req := range requests {
		if err := req.ESAddDocument(); err != nil {
			continue
		}
	}

	return nil
}

func buildRegexpString(str interface{}) string {
	var st []string
	for _, t := range strings.Split(fmt.Sprintf("%v", str), " ") {

		// regexps
		st = append(st, fmt.Sprintf("%v", t))
		st = append(st, fmt.Sprintf("%v.*", t))
		st = append(st, fmt.Sprintf("*.%v", t))
		st = append(st, fmt.Sprintf("(%v)", t))
		// st = append(st, fmt.Sprintf("%v", t))
		// st = append(st, fmt.Sprintf("[%v]", t))
	}

	return strings.Join(st, "|")
}

// CreateContext create context
func CreateContext() context.Context {
	return context.Background()
}
