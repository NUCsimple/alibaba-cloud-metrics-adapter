package vod

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

// AttachedMediaInSearchMedia is a nested struct in vod response
type AttachedMediaInSearchMedia struct {
	Title            string     `json:"Title" xml:"Title"`
	MediaId          string     `json:"MediaId" xml:"MediaId"`
	Ext              string     `json:"Ext" xml:"Ext"`
	CreationTime     string     `json:"CreationTime" xml:"CreationTime"`
	ModificationTime string     `json:"ModificationTime" xml:"ModificationTime"`
	Tags             string     `json:"Tags" xml:"Tags"`
	BusinessType     string     `json:"BusinessType" xml:"BusinessType"`
	URL              string     `json:"URL" xml:"URL"`
	Status           string     `json:"Status" xml:"Status"`
	Description      string     `json:"Description" xml:"Description"`
	StorageLocation  string     `json:"StorageLocation" xml:"StorageLocation"`
	RegionId         string     `json:"RegionId" xml:"RegionId"`
	AppId            string     `json:"AppId" xml:"AppId"`
	Icon             string     `json:"Icon" xml:"Icon"`
	OnlineStatus     string     `json:"OnlineStatus" xml:"OnlineStatus"`
	Categories       []Category `json:"Categories" xml:"Categories"`
}