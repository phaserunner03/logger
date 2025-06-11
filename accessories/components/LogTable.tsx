'use client';

import { LogEntry } from '@/types/logs';
import { StatusBadge } from './StatusBadge';
import { formatDistanceToNow, parseISO, format } from 'date-fns';
import React, { useState } from 'react';

interface LogTableProps {
  logs: LogEntry[];
  lastUpdated: string;
}

export function LogTable({ logs, lastUpdated }: LogTableProps) {
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  const formatTimestamp = (timestamp: { value: string }) => {
    try {
      const date = parseISO(timestamp.value);
      return format(date, 'MMM dd, yyyy HH:mm:ss.SSS zzz');
    } catch (error) {
      console.error('Error parsing timestamp:', error);
      return String(timestamp.value); // Return string version if parsing fails
    }
  };

  const parseJsonString = (jsonStr: string) => {
    try {
      return JSON.parse(jsonStr);
    } catch {
      return jsonStr;
    }
  };

  return (
    <div className="w-full overflow-hidden">
      <div className="flex justify-between items-center mb-4">
        <div>
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">Log Entries</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Project: {logs[0]?.log_name?.split('/')[1] || 'N/A'}
          </p>
        </div>
        <div className="text-right">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Last updated: {formatDistanceToNow(new Date(lastUpdated))} ago
          </p>
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
            Showing {logs.length} entries
          </p>
        </div>
      </div>

      <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700 shadow">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Timestamp</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Severity</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Service</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider whitespace-nowrap">Resource Type</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Message</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
            {logs.map((log) => (
              <React.Fragment key={log.insert_id}>
                <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer" 
                    onClick={() => setExpandedRow(expandedRow === log.insert_id ? null : log.insert_id)}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {formatTimestamp(log.timestamp)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <StatusBadge severity={log.severity} />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {log.service_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {log.resource_type}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                    <div className="truncate max-w-md">
                      {log.text_payload || (log.json_payload !== 'null' ? 'JSON Data' : 'No message')}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setExpandedRow(expandedRow === log.insert_id ? null : log.insert_id);
                      }}
                      className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200"
                    >
                      {expandedRow === log.insert_id ? 'Hide Details' : 'Show Details'}
                    </button>
                  </td>
                </tr>
                {expandedRow === log.insert_id && (
                  <tr className="bg-gray-50 dark:bg-gray-800">
                    <td colSpan={6} className="px-6 py-4">
                      <div className="grid grid-cols-2 gap-6">
                        <div className="space-y-6">
                          <div>
                            <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">Basic Information</h4>
                            <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                              <dt className="text-gray-500 dark:text-gray-400">Insert ID</dt>
                              <dd className="text-gray-900 dark:text-gray-100 font-mono">{log.insert_id}</dd>
                              <dt className="text-gray-500 dark:text-gray-400">Log Name</dt>
                              <dd className="text-gray-900 dark:text-gray-100 break-all">{log.log_name}</dd>
                              <dt className="text-gray-500 dark:text-gray-400">Trace</dt>
                              <dd className="text-gray-900 dark:text-gray-100 font-mono">{log.trace || 'N/A'}</dd>
                              <dt className="text-gray-500 dark:text-gray-400">Span ID</dt>
                              <dd className="text-gray-900 dark:text-gray-100 font-mono">{log.span_id || 'N/A'}</dd>
                            </dl>
                          </div>

                          {log.text_payload && (
                            <div>
                              <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">Text Payload</h4>
                              <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                                {log.text_payload}
                              </pre>
                            </div>
                          )}
                        </div>

                        <div className="space-y-6">
                          <div>
                            <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">Resource Labels</h4>
                            <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                              {JSON.stringify(parseJsonString(log.resource_labels), null, 2)}
                            </pre>
                          </div>

                          {log.json_payload && log.json_payload !== 'null' && (
                            <div>
                              <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">JSON Payload</h4>
                              <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                                {JSON.stringify(parseJsonString(log.json_payload), null, 2)}
                              </pre>
                            </div>
                          )}

                          {log.http_request && log.http_request !== 'null' && (
                            <div>
                              <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">HTTP Request Details</h4>
                              <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                                {JSON.stringify(parseJsonString(log.http_request), null, 2)}
                              </pre>
                            </div>
                          )}

                          {log.labels && log.labels !== '{}' && (
                            <div>
                              <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">Labels</h4>
                              <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                                {JSON.stringify(parseJsonString(log.labels), null, 2)}
                              </pre>
                            </div>
                          )}

                          {log.source_location && log.source_location !== 'null' && (
                            <div>
                              <h4 className="font-semibold mb-2 text-gray-700 dark:text-gray-300">Source Location</h4>
                              <pre className="text-sm bg-white dark:bg-gray-900 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-auto max-h-40">
                                {JSON.stringify(parseJsonString(log.source_location), null, 2)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
