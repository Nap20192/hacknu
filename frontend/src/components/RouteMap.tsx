import React from 'react';
import type { TelemetryData } from '../types/telemetry';

interface RouteMapProps {
  data: TelemetryData | null;
  history: TelemetryData[];
}

export const RouteMap: React.FC<RouteMapProps> = ({ data, history }) => {
  if (!data?.gps) {
    return (
      <div className="route-map">
        <div className="loading">No GPS data</div>
      </div>
    );
  }

  // Calculate bounds for the route
  const positions = history
    .filter(h => h.gps)
    .map(h => ({ lat: h.gps!.lat, lon: h.gps!.lon }));

  const bounds = positions.reduce(
    (acc, pos) => ({
      minLat: Math.min(acc.minLat, pos.lat),
      maxLat: Math.max(acc.maxLat, pos.lat),
      minLon: Math.min(acc.minLon, pos.lon),
      maxLon: Math.max(acc.maxLon, pos.lon),
    }),
    {
      minLat: data.gps.lat,
      maxLat: data.gps.lat,
      minLon: data.gps.lon,
      maxLon: data.gps.lon,
    }
  );

  // const padding = 0.005;
  const latRange = bounds.maxLat - bounds.minLat || 0.01;
  const lonRange = bounds.maxLon - bounds.minLon || 0.01;

  const svgWidth = 400;
  const svgHeight = 300;

  const latToY = (lat: number) => {
    return ((bounds.maxLat - lat) / latRange) * svgHeight;
  };

  const lonToX = (lon: number) => {
    return ((lon - bounds.minLon) / lonRange) * svgWidth;
  };

  return (
    <div className="route-map">
      <h3>Route Map</h3>
      <svg viewBox={`0 0 ${svgWidth} ${svgHeight}`} className="map-svg">
        {/* Background */}
        <rect width={svgWidth} height={svgHeight} fill="#0f0f23" />

        {/* Grid */}
        {[...Array(5)].map((_, i) => {
          const y = (i / 4) * svgHeight;
          return (
            <line
              key={`h-${i}`}
              x1="0"
              y1={y}
              x2={svgWidth}
              y2={y}
              stroke="rgba(255,255,255,0.05)"
              strokeWidth="1"
            />
          );
        })}
        {[...Array(5)].map((_, i) => {
          const x = (i / 4) * svgWidth;
          return (
            <line
              key={`v-${i}`}
              x1={x}
              y1="0"
              x2={x}
              y2={svgHeight}
              stroke="rgba(255,255,255,0.05)"
              strokeWidth="1"
            />
          );
        })}

        {/* Route trail */}
        {positions.length > 1 && (
          <polyline
            points={positions.map(p => `${lonToX(p.lon)},${latToY(p.lat)}`).join(' ')}
            fill="none"
            stroke="rgba(148, 225, 213, 0.3)"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        )}

        {/* Historical positions */}
        {positions.map((pos, i) => {
          const x = lonToX(pos.lon);
          const y = latToY(pos.lat);
          const opacity = i / positions.length;
          return (
            <circle
              key={`pos-${i}`}
              cx={x}
              cy={y}
              r="2"
              fill={`rgba(148, 225, 213, ${opacity * 0.5})`}
            />
          );
        })}

        {/* Current position */}
        <circle
          cx={lonToX(data.gps.lon)}
          cy={latToY(data.gps.lat)}
          r="5"
          fill="none"
          stroke="#5dd6a8"
          strokeWidth="2"
        />
        <circle cx={lonToX(data.gps.lon)} cy={latToY(data.gps.lat)} r="3" fill="#5dd6a8" />

        {/* Speed direction indicator */}
        {data.speed > 5 && (
          <polygon
            points={`${lonToX(data.gps.lon)},${latToY(data.gps.lat) - 8} 
                     ${lonToX(data.gps.lon) - 5},${latToY(data.gps.lat) + 5}
                     ${lonToX(data.gps.lon) + 5},${latToY(data.gps.lat) + 5}`}
            fill="#5dd6a8"
            opacity="0.6"
          />
        )}
      </svg>

      <div className="map-info">
        <div className="info-item">
          <span className="label">Position:</span>
          <span className="value">
            {data.gps.lat.toFixed(4)}°, {data.gps.lon.toFixed(4)}°
          </span>
        </div>
        <div className="info-item">
          <span className="label">Distance:</span>
          <span className="value">{(positions.length * 0.01).toFixed(2)} km</span>
        </div>
      </div>

      <style>{`
        .route-map {
          background: #1a1a2e;
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 8px;
          padding: 16px;
        }

        .route-map h3 {
          margin: 0 0 12px 0;
          font-size: 12px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.5);
          letter-spacing: 0.5px;
        }

        .map-svg {
          width: 100%;
          background: #0f0f23;
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 6px;
          max-height: 300px;
          margin-bottom: 12px;
        }

        .map-info {
          display: flex;
          gap: 16px;
          flex-wrap: wrap;
          font-size: 12px;
        }

        .info-item {
          display: flex;
          gap: 8px;
        }

        .label {
          color: rgba(255, 255, 255, 0.4);
        }

        .value {
          color: rgba(255, 255, 255, 0.8);
          font-weight: 500;
        }

        .loading {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 300px;
          color: rgba(255, 255, 255, 0.3);
          font-size: 14px;
        }
      `}</style>
    </div>
  );
};
