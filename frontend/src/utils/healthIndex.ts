import type { TelemetryData, HealthIndexBreakdown } from '../types/telemetry';

export function calculateHealthIndex(data: TelemetryData): HealthIndexBreakdown {
  const factors = [];

  // Speed factor (optimal 0-120 km/h)
  const speedFactor = calculateSpeedFactor(data.speed);
  factors.push({
    name: 'Speed',
    weight: 0.15,
    value: data.speed,
    impact: speedFactor * 0.15,
  });

  // Fuel factor (optimal > 30%)
  const fuelFactor = calculateFuelFactor(data.fuel.level);
  factors.push({
    name: 'Fuel Level',
    weight: 0.15,
    value: data.fuel.level,
    impact: fuelFactor * 0.15,
  });

  // Fuel consumption factor (< 50 л/ч)
  const consumptionFactor = calculateConsumptionFactor(data.fuel.consumption);
  factors.push({
    name: 'Fuel Consumption',
    weight: 0.1,
    value: data.fuel.consumption,
    impact: consumptionFactor * 0.1,
  });

  // Pressure factors
  const airPressureFactor = calculatePressureFactor(data.pressure.air, 80, 120);
  factors.push({
    name: 'Air Pressure',
    weight: 0.1,
    value: data.pressure.air,
    impact: airPressureFactor * 0.1,
  });

  const brakePressureFactor = calculatePressureFactor(data.pressure.brake, 0, 100);
  factors.push({
    name: 'Brake Pressure',
    weight: 0.1,
    value: data.pressure.brake,
    impact: brakePressureFactor * 0.1,
  });

  // Temperature factors
  const engineTempFactor = calculateTemperatureFactor(data.temperature.engine, 80, 110);
  factors.push({
    name: 'Engine Temperature',
    weight: 0.1,
    value: data.temperature.engine,
    impact: engineTempFactor * 0.1,
  });

  const hydraulicTempFactor = calculateTemperatureFactor(
    data.temperature.hydraulic,
    40,
    80
  );
  factors.push({
    name: 'Hydraulic Temperature',
    weight: 0.08,
    value: data.temperature.hydraulic,
    impact: hydraulicTempFactor * 0.08,
  });

  // Electrical factors
  const voltageFactor = calculateVoltageFactor(data.electrical.voltage);
  factors.push({
    name: 'Voltage',
    weight: 0.1,
    value: data.electrical.voltage,
    impact: voltageFactor * 0.1,
  });

  const batteryFactor = calculateBatteryFactor(data.electrical.batteryHealth);
  factors.push({
    name: 'Battery Health',
    weight: 0.12,
    value: data.electrical.batteryHealth,
    impact: batteryFactor * 0.12,
  });

  // Alerts penalty
  const alertsPenalty = calculateAlertsPenalty(data.alerts);
  factors.push({
    name: 'Alerts Penalty',
    weight: 0,
    value: data.alerts.length,
    impact: -alertsPenalty,
  });

  // Calculate total score
  const score = Math.max(
    0,
    Math.min(
      100,
      factors.reduce((sum, f) => sum + f.impact, 0) + alertsPenalty
    )
  );

  // Sort factors by impact (top 5)
  const topFactors = factors.sort((a, b) => Math.abs(b.impact) - Math.abs(a.impact)).slice(0, 5);

  // Determine category
  let category: 'normal' | 'warning' | 'critical' = 'normal';
  if (score < 50) category = 'critical';
  else if (score < 75) category = 'warning';

  // Get recommendations
  const recommendations = getRecommendations(data, factors, category);

  return {
    score: Math.round(score),
    category,
    factors: topFactors,
    recommendations,
  };
}

function calculateSpeedFactor(speed: number): number {
  if (speed < 0) return 0;
  if (speed <= 120) return 100;
  if (speed <= 150) return 80;
  return 60;
}

function calculateFuelFactor(level: number): number {
  if (level > 50) return 100;
  if (level > 30) return 90;
  if (level > 10) return 70;
  return 40;
}

function calculateConsumptionFactor(consumption: number): number {
  if (consumption < 30) return 100;
  if (consumption < 50) return 90;
  if (consumption < 70) return 70;
  return 50;
}

function calculatePressureFactor(pressure: number, minSafe: number, maxSafe: number): number {
  if (pressure >= minSafe && pressure <= maxSafe) return 100;
  if (Math.abs(pressure - minSafe) < 10 || Math.abs(pressure - maxSafe) < 10) return 85;
  return 60;
}

function calculateTemperatureFactor(temp: number, minSafe: number, maxSafe: number): number {
  if (temp >= minSafe && temp <= maxSafe) return 100;
  if (Math.abs(temp - minSafe) < 5 || Math.abs(temp - maxSafe) < 5) return 85;
  return 60;
}

function calculateVoltageFactor(voltage: number): number {
  if (voltage >= 22 && voltage <= 28) return 100;
  if (voltage >= 20 && voltage <= 30) return 85;
  return 60;
}

function calculateBatteryFactor(health: number): number {
  if (health > 80) return 100;
  if (health > 60) return 85;
  if (health > 40) return 70;
  return 50;
}

function calculateAlertsPenalty(alerts: any[]): number {
  const criticalCount = alerts.filter(a => a.severity === 'critical').length;
  const warningCount = alerts.filter(a => a.severity === 'warning').length;
  return Math.min(40, criticalCount * 20 + warningCount * 5);
}

function getRecommendations(
  data: TelemetryData,
  _factors: any[],
  category: string
): string[] {
  const recommendations: string[] = [];

  if (data.fuel.level < 20) {
    recommendations.push('⚠️ Fuel level critical - refuel as soon as possible');
  }
  if (data.temperature.engine > 100) {
    recommendations.push('🌡️ Engine temperature high - reduce load or cool down');
  }
  if (data.electrical.voltage < 22 || data.electrical.voltage > 28) {
    recommendations.push('⚡ Electrical system abnormal - check alternator');
  }
  if (data.pressure.brake > 100 || data.pressure.brake < 30) {
    recommendations.push('🛑 Brake pressure abnormal - check brake system');
  }
  if (data.alerts.length > 0) {
    recommendations.push('🔔 Multiple alerts detected - review alert details');
  }
  if (category === 'critical') {
    recommendations.push('🚨 System critical - reduce speed and proceed with caution');
  }

  return recommendations.slice(0, 3);
}
