'use client';

import { useStatsModel, useStatsAPIKey } from '@/api/endpoints/stats';
import { useAPIKeyList } from '@/api/endpoints/apikey';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { useMemo } from 'react';
import { Pie, PieChart, Cell } from 'recharts';
import { useTranslations } from 'next-intl';
import { Tabs, TabsList, TabsTrigger } from '@/components/animate-ui/components/animate/tabs';
import { useHomeViewStore, type PieMetricType } from '@/components/modules/home/store';
import { AnimatedNumber } from '@/components/common/AnimatedNumber';
import { formatCount, formatMoney } from '@/lib/utils';

const PIE_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
  'var(--color-chart-6)',
  'var(--color-chart-7)',
  'var(--color-chart-8)',
];

function getPieColor(index: number): string {
  return PIE_COLORS[index % PIE_COLORS.length];
}

function ModelPieChart() {
  const { data: statsModel } = useStatsModel();
  const t = useTranslations('home.pie');
  const pieMetricType = useHomeViewStore((state) => state.modelPieMetric);
  const setPieMetricType = useHomeViewStore((state) => state.setModelPieMetric);

  const getDataKey = (type: PieMetricType) => {
    return type === 'cost' ? 'total_cost' : type === 'count' ? 'request_count' : 'total_token';
  };

  const chartData = useMemo(() => {
    if (!statsModel || statsModel.length === 0) return [];
    const dataKey = getDataKey(pieMetricType);
    return statsModel
      .map((stat) => ({
        name: stat.name || `Model #${stat.id}`,
        value: stat[dataKey].raw,
      }))
      .filter((d) => d.value > 0)
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [statsModel, pieMetricType]);

  const total = useMemo(() => {
    if (!statsModel || statsModel.length === 0) return { requests: 0, cost: 0, tokens: 0 };
    return {
      requests: statsModel.reduce((acc, s) => acc + s.request_count.raw, 0),
      cost: statsModel.reduce((acc, s) => acc + s.total_cost.raw, 0),
      tokens: statsModel.reduce((acc, s) => acc + s.total_token.raw, 0),
    };
  }, [statsModel]);

  const chartConfig = useMemo(() => {
    const config: Record<string, { label: string }> = {};
    chartData.forEach((d) => {
      config[d.name] = { label: d.name };
    });
    return config;
  }, [chartData]);

  if (!chartData.length) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
        <p className="text-sm">{t('noModelData')}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0">
      <div className="flex justify-between items-center px-2">
        <div className="flex gap-2 text-sm">
          {pieMetricType === 'cost' ? (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatMoney(total.cost).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatMoney(total.cost).formatted.unit}</span>
            </div>
          ) : pieMetricType === 'count' ? (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatCount(total.requests).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatCount(total.requests).formatted.unit}</span>
            </div>
          ) : (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatCount(total.tokens).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatCount(total.tokens).formatted.unit}</span>
            </div>
          )}
        </div>
        <Tabs value={pieMetricType} onValueChange={(v) => setPieMetricType(v as PieMetricType)}>
          <TabsList>
            <TabsTrigger value="cost">{t('metricType.cost')}</TabsTrigger>
            <TabsTrigger value="count">{t('metricType.count')}</TabsTrigger>
            <TabsTrigger value="tokens">{t('metricType.tokens')}</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      <ChartContainer config={chartConfig} className="h-48 w-full">
        <PieChart>
          <Pie
            data={chartData}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            outerRadius={70}
            innerRadius={35}
          >
            {chartData.map((entry, index) => (
              <Cell key={entry.name} fill={getPieColor(index)} />
            ))}
          </Pie>
          <ChartTooltip content={<ChartTooltipContent indicator="dot" />} />
        </PieChart>
      </ChartContainer>
      <div className="flex flex-wrap gap-x-3 gap-y-1 px-2 pb-2 justify-center">
        {chartData.slice(0, 8).map((entry, index) => (
          <div key={entry.name} className="flex items-center gap-1 text-xs">
            <div
              className="w-2.5 h-2.5 rounded-full shrink-0"
              style={{ backgroundColor: getPieColor(index) }}
            />
            <span className="text-muted-foreground truncate max-w-[80px]">{entry.name}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function APIKeyPieChart() {
  const { data: statsAPIKey } = useStatsAPIKey();
  const { data: apiKeys } = useAPIKeyList();
  const t = useTranslations('home.pie');
  const pieMetricType = useHomeViewStore((state) => state.apikeyPieMetric);
  const setPieMetricType = useHomeViewStore((state) => state.setApikeyPieMetric);

  const getDataKey = (type: PieMetricType) => {
    return type === 'cost' ? 'total_cost' : type === 'count' ? 'request_count' : 'total_token';
  };

  const nameMap = useMemo(() => {
    if (!apiKeys) return new Map<number, string>();
    const map = new Map<number, string>();
    apiKeys.forEach((key) => map.set(key.id, key.name));
    return map;
  }, [apiKeys]);

  const chartData = useMemo(() => {
    if (!statsAPIKey || statsAPIKey.length === 0) return [];
    const dataKey = getDataKey(pieMetricType);
    return statsAPIKey
      .map((stat) => ({
        name: nameMap.get(stat.api_key_id) || `Key #${stat.api_key_id}`,
        value: stat[dataKey].raw,
      }))
      .filter((d) => d.value > 0)
      .sort((a, b) => b.value - a.value)
      .slice(0, 10);
  }, [statsAPIKey, pieMetricType, nameMap]);

  const total = useMemo(() => {
    if (!statsAPIKey || statsAPIKey.length === 0) return { requests: 0, cost: 0, tokens: 0 };
    return {
      requests: statsAPIKey.reduce((acc, s) => acc + s.request_count.raw, 0),
      cost: statsAPIKey.reduce((acc, s) => acc + s.total_cost.raw, 0),
      tokens: statsAPIKey.reduce((acc, s) => acc + s.total_token.raw, 0),
    };
  }, [statsAPIKey]);

  const chartConfig = useMemo(() => {
    const config: Record<string, { label: string }> = {};
    chartData.forEach((d) => {
      config[d.name] = { label: d.name };
    });
    return config;
  }, [chartData]);

  if (!chartData.length) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
        <p className="text-sm">{t('noAPIKeyData')}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0">
      <div className="flex justify-between items-center px-2">
        <div className="flex gap-2 text-sm">
          {pieMetricType === 'cost' ? (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatMoney(total.cost).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatMoney(total.cost).formatted.unit}</span>
            </div>
          ) : pieMetricType === 'count' ? (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatCount(total.requests).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatCount(total.requests).formatted.unit}</span>
            </div>
          ) : (
            <div className="text-xl font-semibold">
              <AnimatedNumber value={formatCount(total.tokens).formatted.value} />
              <span className="ml-0.5 text-sm text-muted-foreground">{formatCount(total.tokens).formatted.unit}</span>
            </div>
          )}
        </div>
        <Tabs value={pieMetricType} onValueChange={(v) => setPieMetricType(v as PieMetricType)}>
          <TabsList>
            <TabsTrigger value="cost">{t('metricType.cost')}</TabsTrigger>
            <TabsTrigger value="count">{t('metricType.count')}</TabsTrigger>
            <TabsTrigger value="tokens">{t('metricType.tokens')}</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      <ChartContainer config={chartConfig} className="h-48 w-full">
        <PieChart>
          <Pie
            data={chartData}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            outerRadius={70}
            innerRadius={35}
          >
            {chartData.map((entry, index) => (
              <Cell key={entry.name} fill={getPieColor(index)} />
            ))}
          </Pie>
          <ChartTooltip content={<ChartTooltipContent indicator="dot" />} />
        </PieChart>
      </ChartContainer>
      <div className="flex flex-wrap gap-x-3 gap-y-1 px-2 pb-2 justify-center">
        {chartData.slice(0, 8).map((entry, index) => (
          <div key={entry.name} className="flex items-center gap-1 text-xs">
            <div
              className="w-2.5 h-2.5 rounded-full shrink-0"
              style={{ backgroundColor: getPieColor(index) }}
            />
            <span className="text-muted-foreground truncate max-w-[80px]">{entry.name}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function StatsPieCharts() {
  const t = useTranslations('home.pie');

  return (
    <div className="rounded-3xl bg-card border-card-border border text-card-foreground custom-shadow">
      <div className="p-4 pb-0">
        <h3 className="font-semibold text-base mb-3">{t('title')}</h3>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 px-4 pb-4">
        <div className="rounded-2xl bg-accent/5 p-3">
          <h4 className="text-sm font-medium text-muted-foreground mb-2 px-2">{t('modelCalls')}</h4>
          <ModelPieChart />
        </div>
        <div className="rounded-2xl bg-accent/5 p-3">
          <h4 className="text-sm font-medium text-muted-foreground mb-2 px-2">{t('apiKeyUsage')}</h4>
          <APIKeyPieChart />
        </div>
      </div>
    </div>
  );
}
