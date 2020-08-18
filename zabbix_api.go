package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "log"
    "io"
    "net/http"
)

const (
    JsonrpcVersion string = "2.0"
    JsonAuthID int = 112233
)

type ZabbixAPI struct {
    url         string
    user        string
    password    string
    id          int
    auth        string
    Client      *http.Client
}

type JsonRPCRequsetBase struct {
    Jsonrpc     string      `json:"jsonrpc"`
    Method      string      `json:"method"`
    Params      interface{} `json:"params"`
    Id          int         `json:"id"`
}

type JsonRPCRequset struct {
    Jsonrpc     string      `json:"jsonrpc"`
    Method      string      `json:"method"`
    Params      interface{} `json:"params"`
    Id          int         `json:"id"`
    Auth        string      `json:"auth"`
}

type JsonRPCResponse struct {
    Jsonrpc     string          `json:"jsonrpc"`
    Result      interface{}     `json:"result"`
    Error       ZabbixAPIError  `json:"error"`
    Id          int             `json:"id"`
}

type ZabbixAPIError struct {
    Code    int     `json:"code"`
    Message string  `json:"message"`
    Data    string  `json:"data"`
}

type ZHostGroup map[string]interface{}
type ZHost map[string]interface{}
type ZTemplate map[string]interface{}
type ZValuemap map[string]interface{}

func NewZabbixAPI(url, user, password string) (*ZabbixAPI, error) {
    return &ZabbixAPI{
        url: url,
        user: user,
        password: password,
        auth: "",
        id: JsonAuthID,
        Client: &http.Client{},
    }, nil
}


func (api *ZabbixAPI) Request(method string, params interface{}) (JsonRPCResponse, error) {
    id := api.id
    api.id = api.id + 1
    var err error
    var reqJson []byte
    if method != "user.login" {
        reqObj := JsonRPCRequset{
            Jsonrpc: JsonrpcVersion,
            Method: method,
            Params: params,
            Auth: api.auth,
            Id: id,
        }
        reqJson, err = json.Marshal(reqObj)
        if err != nil {
            return JsonRPCResponse{}, err
        }
    } else {
        reqObj := JsonRPCRequsetBase{
            Jsonrpc: JsonrpcVersion,
            Method: method,
            Params: params,
            Id: id,
        }
        reqJson, err = json.Marshal(reqObj)
        if err != nil {
            return JsonRPCResponse{}, err
        }
    }

    log.Println(string(reqJson))

    req, err := http.NewRequest("POST", api.url, bytes.NewBuffer(reqJson))
    if err != nil {
        return JsonRPCResponse{}, err
    }
    req.Header.Add("Content-Type", "application/json-rpc")

    rsp, err := api.Client.Do(req)
    if err != nil {
        return JsonRPCResponse{}, err
    }

    var res JsonRPCResponse
    var buf bytes.Buffer
    _, err = io.Copy(&buf, rsp.Body)
    if err != nil {
        return JsonRPCResponse{}, err
    }
    json.Unmarshal(buf.Bytes(), &res)

    rsp.Body.Close()

    log.Println("res:", res)

    return res, nil
}

func (api *ZabbixAPI) Login() (bool, error) {
    params := make(map[string]string, 0)
    params["user"] = api.user
    params["password"] = api.password

    rsp, err := api.Request("user.login", params)
    if err != nil {
        return false, err
    }
    if rsp.Error.Code != 0 {
        return false, errors.New(rsp.Error.Data)
    }

    api.auth = rsp.Result.(string)
    return true, nil
}

func (api *ZabbixAPI) Logout() (bool, error) {
    params := make(map[string]string, 0)
    rsp, err := api.Request("user.logout", params)
    if err != nil {
        return false, err
    }
    if rsp.Error.Code != 0 {
        return false, errors.New(rsp.Error.Data)
    }

    return true, nil
}

func (api *ZabbixAPI) Host(method string, params interface{}) ([]ZHost, error) {
    rsp, err := api.Request("host."+method, params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    res, err := json.Marshal(rsp.Result)
    var ret []ZHost
    err = json.Unmarshal(res, &ret)
    return ret, nil
}

func (api *ZabbixAPI) HostGroup(method string, params interface{}) ([]ZHostGroup, error) {
    rsp, err := api.Request("hostgroup."+method, params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    res, err := json.Marshal(rsp.Result)
    var ret []ZHostGroup
    err = json.Unmarshal(res, &ret)
    return ret, nil
}

func (api *ZabbixAPI) Template(method string, params interface{}) ([]ZTemplate, error) {
    rsp, err := api.Request("template."+method, params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    res, err := json.Marshal(rsp.Result)
    var ret []ZTemplate
    err = json.Unmarshal(res, &ret)
    return ret, nil
}

func (api *ZabbixAPI) Configuration(method string, params interface{}) (interface{}, error) {
    rsp, err := api.Request("configuration."+method, params)
    if err != nil {
        return "", err
    }
    if rsp.Error.Code != 0 {
        return "", errors.New(rsp.Error.Data)
    }

    res := rsp.Result
    return res, nil
}

func (api *ZabbixAPI) Valuemap(method string, params interface{}) ([]ZValuemap, error) {
    rsp, err := api.Request("valuemap."+method, params)
    if err != nil {
        return nil, err
    }
    if rsp.Error.Code != 0 {
        return nil, errors.New(rsp.Error.Data)
    }

    res, err := json.Marshal(rsp.Result)
    var ret []ZValuemap
    err = json.Unmarshal(res, &ret)
    return res, nil
}