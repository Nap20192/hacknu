import type { Alert, PagedResponse, Response } from '../types/alerts';

const API_BASE_URL = 'http://localhost:8080'; // This should ideally come from an environment variable

export const alertsApi = {
  /**
   * Get active (unresolved) alerts for a locomotive
   */
  getActiveAlerts: async (locomotiveId: string): Promise<PagedResponse<Alert>> => {
    const response = await fetch(
      `${API_BASE_URL}/api/v1/locomotives/${locomotiveId}/alerts`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to fetch active alerts: ${response.statusText}`);
    }

    return response.json();
  },

  /**
   * Get alerts history for a locomotive in a time range
   */
  getAlertsHistory: async (
    locomotiveId: string,
    from?: Date,
    to?: Date
  ): Promise<PagedResponse<Alert>> => {
    const params = new URLSearchParams();

    if (from) {
      params.append('from', from.toISOString());
    }
    if (to) {
      params.append('to', to.toISOString());
    }

    const query = params.toString();
    const url = `${API_BASE_URL}/api/v1/locomotives/${locomotiveId}/alerts/history${
      query ? '?' + query : ''
    }`;

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch alerts history: ${response.statusText}`);
    }

    return response.json();
  },

  /**
   * Acknowledge an alert by ID
   */
  acknowledgeAlert: async (alertId: number): Promise<Response<Alert>> => {
    const response = await fetch(`${API_BASE_URL}/api/v1/alerts/${alertId}/acknowledge`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to acknowledge alert: ${response.statusText}`);
    }

    return response.json();
  },
};
