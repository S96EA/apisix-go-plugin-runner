/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugins

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
)

func init() {
	err := plugin.RegisterPlugin(&SimpleHeaderCheck{})
	if err != nil {
		log.Fatalf("failed to register plugin say: %s", err)
	}
}

// Say is a demo to show how to return data directly instead of proxying
// it to the upstream.
type SimpleHeaderCheck struct {
}

type SimpleHeaderCheckConf struct {
	RedirectUrl    string `json:"redirect_url"`
	GatewayBaseUrl string `json:"gateway_base_url"`
	CallBackPath   string `json:"call_back_path"`
}

func (p *SimpleHeaderCheck) Name() string {
	return "simple_header_check"
}

func (p *SimpleHeaderCheck) ParseConf(in []byte) (interface{}, error) {
	conf := SimpleHeaderCheckConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *SimpleHeaderCheck) checkToken(token string) bool {
	if len(token) == 0 || strings.HasPrefix(token, "invalid") {
		return false
	}
	return true
}

func (p *SimpleHeaderCheck) checkConfig(conf interface{}) bool {
	redirectUrl := conf.(SimpleHeaderCheckConf).RedirectUrl
	gatewayBaseUrl := conf.(SimpleHeaderCheckConf).GatewayBaseUrl
	callbackPath := conf.(SimpleHeaderCheckConf).CallBackPath
	if len(redirectUrl) == 0 || len(gatewayBaseUrl) == 0 || len(callbackPath) == 0 {
		return false
	}
	return true
}

func (p *SimpleHeaderCheck) processCallback(w http.ResponseWriter, r pkgHTTP.Request) {
	// TODO: get code/token... from r
	//token := "aaaaaaaa"
}

func (p *SimpleHeaderCheck) Filter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	if !p.checkConfig(conf) {
		// TODO: think about how to process bad config
		return
	}

	redirectUrl := conf.(SimpleHeaderCheckConf).RedirectUrl
	gatewayBaseUrl := conf.(SimpleHeaderCheckConf).GatewayBaseUrl
	callbackPath := conf.(SimpleHeaderCheckConf).CallBackPath

	if string(r.Path()) == callbackPath {
		p.processCallback(w, r)
		return
	}

	token := r.Header().Get("X-TOKEN")
	if !p.checkToken(token) {
		http.Redirect(w, &http.Request{}, redirectUrl+"?redirect_url="+url.QueryEscape(gatewayBaseUrl+string(r.Path())), http.StatusFound)
		return
	}

	w.Header().Add("X-Resp-A6-Runner", "Go")
	_, err := w.Write([]byte("pass"))
	if err != nil {
		log.Errorf("failed to write: %s", err)
	}
}
