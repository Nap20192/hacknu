import React from 'react';
import type { HealthIndexBreakdown } from '../types/telemetry';

interface HealthIndexProps {
  data: HealthIndexBreakdown;
  isConnected: boolean;
}

export const HealthIndex: React.FC<HealthIndexProps> = ({ data, isConnected }) => {
  const categoryColors = {
    normal: 'from-green-500 to-green-600',
    warning: 'from-yellow-500 to-yellow-600',
    critical: 'from-red-500 to-red-600',
  };

  const categoryLabels = {
    normal: 'NORMAL',
    warning: 'ATTENTION',
    critical: 'CRITICAL',
  };

  const categoryEmoji = {
    normal: '✅',
    warning: '⚠️',
    critical: '🚨',
  };

  return (
    <div className={`health-index ${data.category}`}>
      <div className={`gradient-bg bg-gradient-to-br ${categoryColors[data.category]}`}></div>

      <div className="content">
        <div className="status-indicator">
          <span className="emoji">{categoryEmoji[data.category]}</span>
          <span className="label">{categoryLabels[data.category]}</span>
          {!isConnected && <span className="disconnect">⚠️ NO CONNECTION</span>}
        </div>

        <div className="score-ring">
          <svg viewBox="0 0 36 36" className="circle">
            <circle
              cx="18"
              cy="18"
              r="16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              opacity="0.2"
            />
            <circle
              cx="18"
              cy="18"
              r="16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeDasharray={`${data.score * 1.005} 100.53`}
              transform="rotate(-90 18 18)"
              className="progress"
            />
          </svg>
          <div className="score-text">{data.score}</div>
        </div>

        <div className="factors">
          <h3>Top Factors</h3>
          <div className="factor-list">
            {data.factors.slice(0, 3).map((factor, i) => (
              <div key={i} className="factor-item">
                <div className="factor-name">{factor.name}</div>
                <div className="factor-bar">
                  <div
                    className="factor-fill"
                    style={{
                      width: `${Math.max(0, factor.value)}%`,
                      opacity: 0.7,
                    }}
                  ></div>
                </div>
                <div className="factor-value">{factor.value.toFixed(1)}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="recommendations">
          <h3>Actions</h3>
          <ul>
            {data.recommendations.map((rec, i) => (
              <li key={i}>{rec}</li>
            ))}
          </ul>
        </div>
      </div>

      <style>{`
        .health-index {
          position: relative;
          background: #1a1a2e;
          border-radius: 16px;
          border: 2px solid rgba(255, 255, 255, 0.1);
          overflow: hidden;
          padding: 24px;
          display: flex;
          flex-direction: column;
          gap: 20px;
          min-height: 500px;
        }

        .gradient-bg {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          height: 4px;
          opacity: 0.8;
        }

        .content {
          position: relative;
          z-index: 1;
          display: flex;
          flex-direction: column;
          gap: 24px;
        }

        .status-indicator {
          display: flex;
          align-items: center;
          gap: 12px;
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 1px;
          color: rgba(255, 255, 255, 0.7);
        }

        .emoji {
          font-size: 24px;
        }

        .disconnect {
          margin-left: auto;
          color: #ff6b6b;
          font-weight: bold;
          animation: pulse 1s infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .score-ring {
          position: relative;
          width: 200px;
          height: 200px;
          margin: 0 auto;
        }

        .circle {
          width: 100%;
          height: 100%;
          color: currentColor;
        }

        .progress {
          color: rgba(255, 255, 255, 0.8);
          transition: stroke-dasharray 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
        }

        .health-index.critical .progress {
          color: #ff5252;
        }

        .health-index.warning .progress {
          color: #ffd54f;
        }

        .health-index.normal .progress {
          color: #5dd6a8;
        }

        .score-text {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          font-size: 48px;
          font-weight: bold;
          color: rgba(255, 255, 255, 0.9);
        }

        .factors {
          flex: 1;
        }

        .factors h3,
        .recommendations h3 {
          font-size: 12px;
          text-transform: uppercase;
          color: rgba(255, 255, 255, 0.5);
          margin-bottom: 8px;
          letter-spacing: 0.5px;
        }

        .factor-list {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .factor-item {
          display: grid;
          grid-template-columns: 100px 1fr 50px;
          gap: 8px;
          align-items: center;
          font-size: 12px;
        }

        .factor-name {
          color: rgba(255, 255, 255, 0.6);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .factor-bar {
          background: rgba(255, 255, 255, 0.1);
          height: 6px;
          border-radius: 3px;
          overflow: hidden;
        }

        .factor-fill {
          background: linear-gradient(90deg, #5dd6a8, #ffd54f);
          height: 100%;
          border-radius: 3px;
        }

        .factor-value {
          color: rgba(255, 255, 255, 0.4);
          text-align: right;
          font-size: 11px;
        }

        .recommendations ul {
          list-style: none;
          padding: 0;
          display: flex;
          flex-direction: column;
          gap: 6px;
        }

        .recommendations li {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.6);
          padding-left: 16px;
          position: relative;
        }

        .recommendations li:before {
          content: '→';
          position: absolute;
          left: 0;
          color: rgba(255, 255, 255, 0.3);
        }
      `}</style>
    </div>
  );
};
