interface TimestampValue {
  value: string;
}

interface TimestampValue {
  value: string;
}

export interface LogEntry {
  timestamp: TimestampValue;
  severity: string;
  log_name: string;
  text_payload: string;
  json_payload: string | null;
  insert_id: string;
  resource_type: string;
  resource_labels: string;
  http_request: string;
  trace: string;
  span_id: string;
  source_location: string | null;
  labels: string;
  service_name: string;
}

export interface LogsResponse {
  logs: LogEntry[];
  lastUpdated: string;
}
