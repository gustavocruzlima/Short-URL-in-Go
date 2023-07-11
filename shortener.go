package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/itchyny/base58-go"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Url struct {
	Url string `json:"url" binding:"required"`
}

type ShortUrl struct {
	Short string `json:"short" binding:"required"`
}

type Short struct {
	Url   string `json:"url" binding:"required"`
	Short string `json:"short" binding:"required"`
}

const uri = "<YOUR-CONNECT-STRING>"

func main() {
	router := gin.Default()
	// send an URL, verify if already does not exists and create a short version of it
	router.POST("/short", shortUrl)

	// send a shorURL and receive its full url if exists
	router.POST("/search", SearchURL)

	router.Run()

}

func connectMongo() (*mongo.Collection, *mongo.Client, error) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)

	//connect to collection shorturl
	coll := client.Database("shortener_db").Collection("shorturl")

	return coll, client, err

}

func SearchData(coll *mongo.Collection, client *mongo.Client, err error, searchType string, data string) (string, string) {

	//verify errors on connection
	if err != nil {
		panic(err)
	}

	//verify what kind of search and return
	if searchType == "url" {
		urlData := strings.Split(data, "Url")

		filter := bson.D{{"url", urlData[0]}}

		var result Short
		err = coll.FindOne(context.TODO(), filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// This error means your query did not match any documents.
				fmt.Println("Error: " + err.Error())
				return "", ""
			}
		}

		return result.Url, result.Short
	} else {
		shortData := strings.Split(data, "Short")

		filter := bson.D{{"short", shortData[0]}}

		var result Short
		err = coll.FindOne(context.TODO(), filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// This error means your query did not match any documents.
				fmt.Println("Error: " + err.Error())
				return "", ""
			}
		}

		return result.Url, result.Short
	}

}

func CreateShortLink(Link string) string {
	urlSHA := EncryptSHA256(Link)
	SHACrypt := new(big.Int).SetBytes(urlSHA).Uint64()
	ShortLink := base58Encoded([]byte(fmt.Sprintf("%d", SHACrypt)))
	return ShortLink[:8]
}

func EncryptSHA256(Link string) []byte {
	crypt := sha256.New()
	crypt.Write([]byte(Link))
	return crypt.Sum(nil)
}

func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return string(encoded)
}

func shortUrl(c *gin.Context) {

	url := Url{}
	// binding to URL
	if err := c.ShouldBindJSON(&url); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// create conection
	coll, client, err := connectMongo()

	// get value of URL
	UrlValue := strings.Split(url.Url, "Url")

	// verify if URL already have a short version
	resultUrl, resultShort := SearchData(coll, client, err, "url", UrlValue[0])

	if resultShort != "" {

		//disconnect from mongo
		client.Disconnect(context.TODO())

		c.JSON(200, gin.H{
			"message":  "Url already shortened",
			"url":      resultUrl,
			"shortUrl": resultShort,
		})

	} else {
		ShortVersion := CreateShortLink(url.Url)
		// doc to insert
		doc := Short{Url: url.Url, Short: "shorturl.com/" + ShortVersion}

		//insert Url and shortUrl in collection
		result, err := coll.InsertOne(context.TODO(), doc)

		//disconnect from mongo
		client.Disconnect(context.TODO())

		if err == nil {

			fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)

			c.JSON(200, gin.H{
				"message": "Url shortened successfully",
				"url":     url.Url,
				"short":   "shorturl.com/" + ShortVersion,
			})
		} else {
			c.JSON(500, gin.H{
				"message": "error during the creation of the short url",
			})
		}

	}
}

func SearchURL(c *gin.Context) {
	shortURL := ShortUrl{}
	// binding to URL
	if err := c.ShouldBindJSON(&shortURL); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// create conection
	coll, client, err := connectMongo()

	// get value of short URL
	shortUrlValue := strings.Split(shortURL.Short, "Short")

	// search short URL
	resultUrl, resultShort := SearchData(coll, client, err, "short", shortUrlValue[0])

	client.Disconnect(context.TODO())

	if resultShort == "" {
		c.JSON(404, gin.H{
			"message": "Url not found!",
		})
	} else {
		c.JSON(200, gin.H{
			"url":      resultUrl,
			"shortUrl": resultShort,
		})
	}

}
