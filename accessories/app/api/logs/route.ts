import { BigQuery } from '@google-cloud/bigquery';
import { NextResponse } from 'next/server';
import path from 'path';

const keyFilePath = path.resolve(process.cwd(), '../key.json');

const bigquery = new BigQuery({
  keyFilename: keyFilePath,
  projectId: 'logger-462111',
});

export async function GET() {
  try {
    // Test the connection first
    await bigquery.dataset('logs').exists();

    const query = `
      SELECT 
        timestamp,
        severity,
        log_name,
        COALESCE(text_payload, '') as text_payload,
        COALESCE(TO_JSON_STRING(json_payload), 'null') as json_payload,
        insert_id,
        resource_type,
        TO_JSON_STRING(resource_labels) as resource_labels,
        TO_JSON_STRING(http_request) as http_request,
        trace,
        span_id,
        TO_JSON_STRING(source_location) as source_location,
        TO_JSON_STRING(labels) as labels,
        service_name
      FROM \`logger-462111.logging.logs_table\`
      ORDER BY timestamp DESC
      LIMIT 100
    `;

    const [rows] = await bigquery.query({ query });

    return NextResponse.json({
      logs: rows,
      lastUpdated: new Date().toISOString(),
    });
  } catch (error) {
    console.error('Error fetching logs:', error);
    return NextResponse.json(
      { 
        error: 'Failed to fetch logs',
        details: error instanceof Error ? error.message : 'Unknown error'
      },
      { status: 500 }
    );
  }
}
