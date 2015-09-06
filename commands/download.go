package commands

import (
  "strings"
  "encoding/json"
  "github.com/JFrogDev/artifactory-cli-go/utils"
)

var MinConcurrentDownloadSize = 1 // 10240000

func Download(url string, downloadPattern string, recursive bool, props string, user string, password string, flat bool, dryRun bool) {
    aqlUrl := url + "api/search/aql"
    if strings.HasSuffix(downloadPattern, "/") {
        downloadPattern += "*"
    }

    data := utils.BuildAqlSearchQuery(downloadPattern, recursive, props)

    println("AQL query: " + data)

    json := utils.SendPost(aqlUrl, []byte(data), user, password)
    resultItems := parseAqlSearchResponse(json)
    downloadFiles(resultItems, url, user, password, flat, dryRun)
}

func downloadFiles(resultItems []AqlSearchResultItem, url string, user string, password string, flat bool, dryRun bool) {
    size := len(resultItems)

    for i := 0; i < size; i++ {
        downloadPath := buildDownloadUrl(url, resultItems[i])
        print("Downloading " + downloadPath + "...")

        if !dryRun {
            details := utils.GetFileDetailsFromArtifactory(downloadPath, user, password)
            localFilePath := resultItems[i].Path + "/" + resultItems[i].Name
            if shouldDownloadFile(localFilePath, details, user, password) {
                if !details.AcceptRanges || details.Size < MinConcurrentDownloadSize {
                    resp := utils.DownloadFile(downloadPath, resultItems[i].Path, resultItems[i].Name, flat, user, password)
                    println("Artifactory response:", resp.Status)
                } else {
                    utils.DownloadFileConcurrently(downloadPath, resultItems[i].Path, resultItems[i].Name, flat, user, password, details.Size)
                }
            } else {
                println("File already exists locally.")
            }
        }
    }
}

func buildDownloadUrl(baseUrl string, resultItem AqlSearchResultItem) string {
    if resultItem.Path == "." {
        return baseUrl + resultItem.Repo + "/" + resultItem.Name
    }
    return baseUrl + resultItem.Repo + "/" + resultItem.Path + "/" + resultItem.Name
}

func shouldDownloadFile(localFilePath string, artifactoryFileDetails *utils.FileDetails, user string, password string) bool {
    if !utils.IsFileExists(localFilePath) {
        return true
    }
    localFileDetails := utils.GetFileDetails(utils.ReadFile(localFilePath))
    if localFileDetails.Md5 != artifactoryFileDetails.Md5 || localFileDetails.Sha1 != artifactoryFileDetails.Sha1 {
       return true
    }
    return false
}

func parseAqlSearchResponse(resp []byte) []AqlSearchResultItem {
    var result AqlSearchResult
    err := json.Unmarshal(resp, &result)

    utils.CheckError(err)
    return result.Results
}

type AqlSearchResult struct {
    Results []AqlSearchResultItem
}

type AqlSearchResultItem struct {
     Repo string
     Path string
     Name string
 }