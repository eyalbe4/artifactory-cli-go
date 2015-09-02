package utils

import (
  "strings"
)

func BuildAqlSearchQuery(searchPattern string, recursive bool, props string) string {
    index := strings.Index(searchPattern, "/")
    if index == -1 {
        Exit("Invalid search pattern: " + searchPattern)
    }

    repo := searchPattern[:index]
    searchPattern = searchPattern[index+1:]

    pairs := createPathFilePairs(searchPattern, recursive)
    size := len(pairs)

    json :=
        "{" +
            "\"repo\": \"" + repo + "\"," +
            buildPropsQuery(props) +
            "\"$or\": ["

    if size == 0 {
        json +=
            "{" +
                buildInnerQuery(repo, ".", searchPattern) +
            "}"
    } else {
        for i := 0; i < size; i++ {
            json +=
                "{" +
                    buildInnerQuery(repo, pairs[i].path, pairs[i].file) +
                "}"

            if (i+1 < size) {
                json += ","
            }
        }
    }

    json +=
            "]" +
        "}"

    return "items.find(" + json + ")"
}

func buildPropsQuery(props string) string {
    if props == "" {
        return ""
    }
    propList := strings.Split(props, ";")
    query := ""
    for _, prop := range propList {
        keyVal := strings.Split(prop, "=")
        if len(keyVal) != 2 {
            Exit("Invalid props pattern: " + props)
        }
        key := keyVal[0]
        value := keyVal[1]
        query +=
            "\"@" + key + "\": {\"$match\" : \"" + value + "\"},"
    }
    return query
}

func buildInnerQuery(repo string, path string, name string) string {
    query :=
        "\"$and\": [{" +
            "\"path\": {" +
                "\"$match\":" + "\"" +  path + "\"" +
            "}," +
            "\"name\":{" +
                "\"$match\":" + "\"" + name + "\"" +
            "}" +
        "}]"

    return query
}

// We need to translate the provided download pattern to an AQL query.
// In Artifactory, for each artifact the name and path of the artifact are saved separately.
// We therefore need to build an AQL query that covers all possible paths and names the provided
// pattern can include.
// For example, the pattern a/* can include the two following files:
// a/file1.tgz and also a/b/file2.tgz
// To achieve that, this function parses the pattern by splitting it by its * characters.
// The end result is a list of PathFilePair structs.
// Each struct represent a possible path and file name pair to be included in AQL query with an "or" relationship.
func createPathFilePairs(pattern string, recursive bool) []PathFilePair {
    pairs := []PathFilePair{}
    if (pattern == "*") {
        if recursive {
            pairs = append(pairs, PathFilePair{"*", "*"})
        } else {
            pairs = append(pairs, PathFilePair{".", "*"})
        }
        return pairs
    }

    Index := strings.LastIndex(pattern, "/")
    path := ""
    if (Index > 0) {
        path = pattern[0:Index]
        name := pattern[Index+1:]
        pairs = append(pairs, PathFilePair{path, name})
        if !recursive {
            return pairs
        }
        if name == "*" {
            path += "/*"
            pairs = append(pairs, PathFilePair{path, "*"})
            return pairs
        }
        pattern = name
    }

    sections := strings.Split(pattern, "*")
    size := len(sections)

    for i := 0; i < size; i++ {
        if (sections[i] == "") {
            continue
        }
        options := []string{}
        if (i > 0) {
            options = append(options, "/" + sections[i])
        }
        if (i+1 < size) {
            options = append(options, sections[i] + "/")
        }
        for _, option := range options {
            str := ""
            for j := 0; j < size; j++ {
                if (j > 0) {
                    str += "*"
                }
                if (j == i) {
                    str += option
                } else {
                    str += sections[j]
                }
            }
            split := strings.Split(str, "/")
            pairs = append(pairs, PathFilePair{path + split[0], split[1]})
        }
    }
    return pairs
}

type PathFilePair struct {
    path string
    file string
}