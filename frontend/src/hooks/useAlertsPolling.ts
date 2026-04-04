import { useEffect, useRef, useState } from 'react';
import { alertsApi } from '../api/alerts';
import type { AlertsState, LongPollingConfig } from '../types/alerts';

/**
 * Custom hook for long polling alerts
 * Automatically fetches active alerts for a locomotive at regular intervals
 */
export const useAlertsPolling = (
  locomotiveId: string | null,
  config: LongPollingConfig = {}
) => {
  const { interval = 3000, timeout = 5000, maxRetries = 5 } = config;

  const [state, setState] = useState<AlertsState>({
    active: [],
    loading: true,
    error: null,
    lastUpdated: null,
  });

  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);
  const abortControllerRef = useRef<AbortController | null>(null);
  const retryCountRef = useRef(0);

  useEffect(() => {
    if (!locomotiveId) {
      setState({
        active: [],
        loading: false,
        error: null,
        lastUpdated: null,
      });
      return;
    }

    const fetchAlerts = async () => {
      try {
        // Create abort controller for this request
        abortControllerRef.current = new AbortController();
        const timeoutId = setTimeout(() => abortControllerRef.current?.abort(), timeout);

        const response = await alertsApi.getActiveAlerts(locomotiveId);

        clearTimeout(timeoutId);

        setState({
          active: response.data,
          loading: false,
          error: null,
          lastUpdated: new Date(),
        });

        retryCountRef.current = 0;
      } catch (error) {
        retryCountRef.current += 1;

        const errorMessage =
          error instanceof Error ? error.message : 'Unknown error occurred';

        setState((prev) => ({
          ...prev,
          loading: false,
          error: errorMessage,
        }));

        // Stop polling if max retries exceeded
        if (retryCountRef.current >= maxRetries) {
          console.error(`Polling stopped after ${maxRetries} retries`);
          if (pollingIntervalRef.current) {
            clearInterval(pollingIntervalRef.current);
          }
        }
      }
    };

    // Initial fetch
    fetchAlerts();

    // Setup polling interval
    pollingIntervalRef.current = setInterval(fetchAlerts, interval);

    // Cleanup
    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [locomotiveId, interval, timeout, maxRetries]);

  const acknowledgeAlert = async (alertId: number) => {
    try {
      const response = await alertsApi.acknowledgeAlert(alertId);
      if (response.success) {
        // Update local state
        setState((prev) => ({
          ...prev,
          active: prev.active.map((alert) =>
            alert.id === alertId ? { ...alert, acknowledged: true } : alert
          ),
        }));
        return true;
      }
      return false;
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
      return false;
    }
  };

  const stopPolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
    }
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  };

  return {
    ...state,
    acknowledgeAlert,
    stopPolling,
  };
};
