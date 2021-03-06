/*
 * @Copyright Reserved By Janusec (https://www.janusec.com/).
 * @Author: U2
 * @Date: 2018-07-14 16:35:23
 * @Last Modified: U2, 2018-07-14 16:35:23
 */

package firewall

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/Janusec/janusec/data"
	"github.com/Janusec/janusec/models"
	"github.com/Janusec/janusec/utils"
)

func InitHitLog() {
	if data.IsMaster {
		data.DAL.CreateTableIfNotExistsGroupHitLog()
	}
}

func LogGroupHitRequest(r *http.Request, appID int64, clientIP string, policy *models.GroupPolicy) {
	requestTime := time.Now().Unix()
	contentType := r.Header.Get("Content-Type")
	cookies := r.Header.Get("Cookie")
	rawRequestBytes, err := httputil.DumpRequest(r, true)
	utils.CheckError("LogGroupHitRequest DumpRequest", err)
	maxRawSize := len(rawRequestBytes)
	if maxRawSize > 16384 {
		maxRawSize = 16384
	}
	rawRequest := string(rawRequestBytes[:maxRawSize])
	if data.IsMaster {
		data.DAL.InsertGroupHitLog(requestTime, clientIP, r.Host, r.Method, r.URL.Path, r.URL.RawQuery, contentType, r.UserAgent(), cookies, rawRequest, int64(policy.Action), policy.ID, policy.VulnID, appID)
	} else {
		regexHitLog := &models.GroupHitLog{
			RequestTime: requestTime,
			ClientIP:    clientIP,
			Host:        r.Host,
			Method:      r.Method,
			UrlPath:     r.URL.Path,
			UrlQuery:    r.URL.RawQuery,
			ContentType: contentType,
			UserAgent:   r.UserAgent(),
			Cookies:     cookies,
			RawRequest:  rawRequest,
			Action:      policy.Action,
			PolicyID:    policy.ID,
			VulnID:      policy.VulnID,
			AppID:       appID}
		RPCGroupHitLog(regexHitLog)
	}
}

func LogGroupHitRequestAPI(r *http.Request) error {
	var regexHitLogReq models.RPCGroupHitLogRequest
	err := json.NewDecoder(r.Body).Decode(&regexHitLogReq)
	defer r.Body.Close()
	utils.CheckError("LogGroupHitRequestAPI Decode", err)
	regexHitLog := regexHitLogReq.Object
	if regexHitLog == nil {
		return errors.New("LogGroupHitRequestAPI parse body null.")
	}
	return data.DAL.InsertGroupHitLog(regexHitLog.RequestTime, regexHitLog.ClientIP, regexHitLog.Host, regexHitLog.Method, regexHitLog.UrlPath, regexHitLog.UrlQuery, regexHitLog.ContentType, regexHitLog.UserAgent, regexHitLog.Cookies, regexHitLog.RawRequest, int64(regexHitLog.Action), regexHitLog.PolicyID, regexHitLog.VulnID, regexHitLog.AppID)
}

func GetGroupLogCount(param map[string]interface{}) (*models.GroupHitLogsCount, error) {
	appID := int64(param["app_id"].(float64))
	startTime := int64(param["start_time"].(float64))
	endTime := int64(param["end_time"].(float64))
	count, err := data.DAL.SelectGroupHitLogsCount(appID, startTime, endTime)
	logsCount := &models.GroupHitLogsCount{AppID: appID, StartTime: startTime, EndTime: endTime, Count: count}
	return logsCount, err
}

func GetVulnStat(param map[string]interface{}) (vulnStat []*models.VulnStat, err error) {
	appID := int64(param["app_id"].(float64))
	startTime := int64(param["start_time"].(float64))
	endTime := int64(param["end_time"].(float64))
	if appID == 0 {
		vulnStat, err = data.DAL.SelectAllVulnStat(startTime, endTime)
	} else {
		vulnStat, err = data.DAL.SelectVulnStatByAppID(appID, startTime, endTime)
	}
	return vulnStat, err
}

func GetWeekStat(param map[string]interface{}) (weekStat []int64, err error) {
	appID := int64(param["app_id"].(float64))
	vulnID := int64(param["vuln_id"].(float64))
	startTime := int64(param["start_time"].(float64))
	for i := int64(0); i < 7; i++ {
		dayStartTime := startTime + 86400*i
		dayEndTime := dayStartTime + 86400
		if appID == 0 {
			if vulnID == 0 {
				dayCount, err := data.DAL.SelectAllGroupHitLogsCount(dayStartTime, dayEndTime)
				utils.CheckError("GetWeekStat SelectAllGroupHitLogsCount", err)
				weekStat = append(weekStat, dayCount)
			} else {
				dayCount, err := data.DAL.SelectAllGroupHitLogsCountByVulnID(vulnID, dayStartTime, dayEndTime)
				utils.CheckError("GetWeekStat SelectAllGroupHitLogsCountByVulnID", err)
				weekStat = append(weekStat, dayCount)
			}

		} else {
			if vulnID == 0 {
				dayCount, err := data.DAL.SelectGroupHitLogsCount(appID, dayStartTime, dayEndTime)
				utils.CheckError("GetWeekStat SelectGroupHitLogsCount", err)
				weekStat = append(weekStat, dayCount)
			} else {
				dayCount, err := data.DAL.SelectGroupHitLogsCountByVulnID(appID, vulnID, dayStartTime, dayEndTime)
				utils.CheckError("GetWeekStat SelectGroupHitLogsCountByVulnID", err)
				weekStat = append(weekStat, dayCount)
			}
		}
	}
	return weekStat, nil
}

func GetGroupLogs(param map[string]interface{}) ([]*models.SimpleGroupHitLog, error) {
	appID := int64(param["app_id"].(float64))
	startTime := int64(param["start_time"].(float64))
	endTime := int64(param["end_time"].(float64))
	requestCount := int64(param["request_count"].(float64))
	offset := int64(param["offset"].(float64))
	simpleRegexHitLogs := data.DAL.SelectGroupHitLogs(appID, startTime, endTime, requestCount, offset)
	return simpleRegexHitLogs, nil
}

func GetGroupLogByID(id int64) (*models.GroupHitLog, error) {
	regexHitLog, err := data.DAL.SelectGroupHitLogByID(id)
	return regexHitLog, err
}
