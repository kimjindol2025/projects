---
layout: post
title: Phase3-017-Logging-ELK-Stack
date: 2026-03-28
---
# 로깅 시스템: ELK Stack으로 100GB 로그 처리하기

## 요약

**배우는 내용**:
- ELK Stack: Elasticsearch + Logstash + Kibana
- 1일 100GB 로그 수집 및 저장
- 실시간 검색 및 분석
- 알림 규칙 설정

---

## 1. ELK Stack 아키텍처

```
┌─────────────────────────────────────┐
│ Application Servers                 │
│ - API Server                        │
│ - Web Server                        │
│ - Database                          │
└──────────────┬──────────────────────┘
               │ 로그 전송
┌──────────────▼──────────────────────┐
│ Logstash (Log Processing)           │
│ - 수집 (beats, syslog)              │
│ - 필터링 (grok, dissect)            │
│ - 변환 (date, mutate)               │
└──────────────┬──────────────────────┘
               │ 인덱싱
┌──────────────▼──────────────────────┐
│ Elasticsearch (저장소)               │
│ - 분산 인덱싱                        │
│ - 전전 검색                         │
│ - 집계 (aggregations)               │
└──────────────┬──────────────────────┘
               │ 조회
┌──────────────▼──────────────────────┐
│ Kibana (시각화)                     │
│ - 대시보드                          │
│ - 그래프                            │
│ - 알림                              │
└─────────────────────────────────────┘
```

---

## 2. Logstash 설정

### 기본 파이프라인

```yaml
# logstash.conf
input {
  # 1. 로그 수집 (Filebeat에서)
  beats {
    port => 5044
  }

  # 2. Syslog 수신
  syslog {
    port => 5000
  }
}

filter {
  # 1. Grok 패턴으로 파싱
  grok {
    match => {
      "message" => "%{COMBINEDAPACHELOG}"
    }
  }

  # 2. 타임스탬프 처리
  date {
    match => ["timestamp", "dd/MMM/yyyy:HH:mm:ss Z"]
    target => "@timestamp"
  }

  # 3. IP 지리정보 추가
  geoip {
    source => "clientip"
  }

  # 4. 필드 변환
  mutate {
    convert => {
      "bytes" => "integer"
      "response" => "integer"
    }
    rename => {
      "response" => "http_status"
    }
  }
}

output {
  # Elasticsearch로 저장
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "logs-%{+YYYY.MM.dd}"
  }

  # 콘솔 출력 (디버깅)
  stdout { codec => rubydebug }
}
```

### 성능 최적화

```yaml
# 배치 처리
output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "logs-%{+YYYY.MM.dd}"

    # 배치 설정
    bulk_size => 1000        # 1000개씩 일괄 전송
    flush_interval => 2      # 2초마다 플러시
    idle_flush_time => 1     # 1초 유휴 시 전송
  }
}
```

---

## 3. Elasticsearch 설정

### 인덱스 설정

```json
{
  "settings": {
    "number_of_shards": 5,
    "number_of_replicas": 1,
    "refresh_interval": "30s"
  },
  "mappings": {
    "properties": {
      "timestamp": {
        "type": "date"
      },
      "message": {
        "type": "text"
      },
      "level": {
        "type": "keyword"
      },
      "http_status": {
        "type": "integer"
      },
      "response_time": {
        "type": "float"
      },
      "location": {
        "type": "geo_point"
      }
    }
  }
}
```

### 성능 팁

```bash
# 1. 대량 수집 중 refresh 비활성화
PUT logs-*/_settings
{
  "refresh_interval": "-1"  # 비활성화
}

# 2. 수집 완료 후 활성화
PUT logs-*/_settings
{
  "refresh_interval": "30s"  # 정상화
}

# 3. 자동 정리 (ILM)
PUT _ilm/policy/logs-policy
{
  "phases": {
    "hot": {
      "min_age": "0d",
      "actions": {
        "rollover": {
          "max_size": "50GB"
        }
      }
    },
    "warm": {
      "min_age": "7d",
      "actions": {
        "set_priority": {
          "priority": 50
        }
      }
    },
    "cold": {
      "min_age": "30d",
      "actions": {
        "set_priority": {
          "priority": 0
        }
      }
    },
    "delete": {
      "min_age": "90d",
      "actions": {
        "delete": {}
      }
    }
  }
}
```

---

## 4. Kibana 분석

### 대시보드 예시

```javascript
// 요청별 응답시간 분포
GET logs-*/_search
{
  "aggs": {
    "response_time_percentiles": {
      "percentiles": {
        "field": "response_time",
        "percents": [50, 90, 95, 99]
      }
    }
  }
}

// 시간별 에러율
GET logs-*/_search
{
  "aggs": {
    "time_buckets": {
      "date_histogram": {
        "field": "timestamp",
        "interval": "1h"
      },
      "aggs": {
        "error_rate": {
          "filter": {
            "range": {
              "http_status": {
                "gte": 400
              }
            }
          }
        }
      }
    }
  }
}

// 지역별 트래픽
GET logs-*/_search
{
  "aggs": {
    "by_location": {
      "geohash_grid": {
        "field": "location",
        "precision": 4
      },
      "aggs": {
        "total_requests": {
          "value_count": {
            "field": "request_id"
          }
        }
      }
    }
  }
}
```

---

## 5. 성능 벤치마크

### 수집 처리량

```
시나리오: 1일 100GB 로그

처리량:
- Logstash: 50K 이벤트/초
- Elasticsearch: 40K 이벤트/초
- 실제: 30K 이벤트/초 (안전 마진)

계산:
- 1일 = 86,400초
- 30K × 86,400 = 2.59억 이벤트
- 평균 350B/이벤트 = 91GB (목표 100GB와 유사)
```

### 저장 공간

```
원본 로그: 100GB
Elasticsearch 저장:
  - 1개 Shard: 100GB
  - 1개 Replica: +100GB
  - 메타데이터: +10GB
  - 총: 210GB

압축 (ILM):
  - Hot (7일): 100GB
  - Warm (23일): 50GB (압축)
  - Cold (60일): 10GB (cold storage)
  - 총: 160GB
```

---

## 6. 알림 규칙

```yaml
# Watcher 알림
PUT _watcher/watch/high_error_rate
{
  "trigger": {
    "schedule": {
      "interval": "5m"
    }
  },
  "input": {
    "search": {
      "request": {
        "indices": ["logs-*"],
        "body": {
          "query": {
            "range": {
              "timestamp": {
                "gte": "now-5m"
              }
            }
          },
          "aggs": {
            "error_rate": {
              "filter": {
                "range": {
                  "http_status": { "gte": 500 }
                }
              }
            }
          }
        }
      }
    }
  },
  "condition": {
    "compare": {
      "ctx.payload.aggregations.error_rate.doc_count": {
        "gt": 100
      }
    }
  },
  "actions": {
    "send_alert": {
      "email": {
        "to": "alerts@example.com",
        "subject": "High error rate detected"
      }
    }
  }
}
```

---

## 7. 문제 해결

### 문제 1: 느린 검색

```
원인: 너무 많은 인덱스 검색
해결:
- 인덱스 제한: logs-2026.03*
- 필터 먼저: 좁은 범위 쿼리

# ❌ 느림
GET logs-*/_search
{
  "query": {
    "match": { "message": "error" }
  }
}

# ✅ 빠름
GET logs-2026.03*/_search
{
  "query": {
    "bool": {
      "must": [
        { "range": { "timestamp": { "gte": "2026-03-01" } } },
        { "match": { "message": "error" } }
      ]
    }
  }
}
```

### 문제 2: 메모리 부족

```
원인: Elasticsearch 메모리 부족
해결:
1. Heap 크기 조정 (최대 32GB)
2. Shard 수 줄이기
3. 오래된 인덱스 삭제 (ILM)
```

---

## 핵심 정리

| 단계 | 도구 | 역할 |
|------|------|------|
| **수집** | Logstash | 로그 처리 |
| **저장** | Elasticsearch | 검색 가능 저장 |
| **분석** | Kibana | 시각화 및 알림 |

---

## 결론

ELK Stack은 **로그 처리의 표준**입니다.

- 1일 100GB 처리 가능
- 실시간 분석
- 자동 알림

🚀 로그로 시스템을 이해하세요!

---

질문이나 피드백은 댓글로 남겨주세요! 💬
