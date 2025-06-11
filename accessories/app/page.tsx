import { LogTable } from '@/components/LogTable';
import { LogsResponse } from '@/types/logs';

async function getLogs(): Promise<LogsResponse> {
  try {
    const res = await fetch('http://localhost:3000/api/logs', {
      next: { revalidate: 30 },
      cache: 'no-store'
    });
    
    if (!res.ok) {
      const errorData = await res.json();
      throw new Error(errorData.details || 'Failed to fetch logs');
    }

    return res.json();
  } catch (error) {
    console.error('Error fetching logs:', error);
    throw error;
  }
}

export default async function Home() {
  try {
    const { logs, lastUpdated } = await getLogs();

    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <main className="container mx-auto px-4 py-8">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="mb-8">
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                Log Anomaly Detection Dashboard
              </h1>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
                Project: {process.env.NEXT_PUBLIC_PROJECT_ID}
              </p>
            </div>
            <LogTable logs={logs} lastUpdated={lastUpdated} />
          </div>
        </main>
      </div>
    );
  } catch (error) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-md w-full">
          <h1 className="text-2xl font-bold text-red-600 dark:text-red-400 mb-4">
            Error Loading Logs
          </h1>
          <p className="text-gray-600 dark:text-gray-300">
            {error instanceof Error ? error.message : 'An unexpected error occurred'}
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-4">
            Please check your BigQuery connection and try again.
          </p>
        </div>
      </div>
    );
  }
}
