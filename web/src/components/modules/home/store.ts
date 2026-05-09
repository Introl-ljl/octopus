'use client';

import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

export type RankSortMode = 'cost' | 'count' | 'tokens';
export type ChartMetricType = 'cost' | 'count' | 'tokens';
export type ChartPeriod = '1' | '7' | '30';
export type PieMetricType = 'cost' | 'count' | 'tokens';

interface HomeViewState {
    rankSortMode: RankSortMode;
    chartMetricType: ChartMetricType;
    chartPeriod: ChartPeriod;
    modelPieMetric: PieMetricType;
    apikeyPieMetric: PieMetricType;
    setRankSortMode: (value: RankSortMode) => void;
    setChartMetricType: (value: ChartMetricType) => void;
    setChartPeriod: (value: ChartPeriod) => void;
    setModelPieMetric: (value: PieMetricType) => void;
    setApikeyPieMetric: (value: PieMetricType) => void;
}

export const useHomeViewStore = create<HomeViewState>()(
    persist(
        (set) => ({
            rankSortMode: 'cost',
            chartMetricType: 'cost',
            chartPeriod: '1',
            modelPieMetric: 'cost',
            apikeyPieMetric: 'cost',
            setRankSortMode: (value) => set({ rankSortMode: value }),
            setChartMetricType: (value) => set({ chartMetricType: value }),
            setChartPeriod: (value) => set({ chartPeriod: value }),
            setModelPieMetric: (value) => set({ modelPieMetric: value }),
            setApikeyPieMetric: (value) => set({ apikeyPieMetric: value }),
        }),
        {
            name: 'home-view-options-storage',
            storage: createJSONStorage(() => localStorage),
            partialize: (state) => ({
                rankSortMode: state.rankSortMode,
                chartMetricType: state.chartMetricType,
                chartPeriod: state.chartPeriod,
                modelPieMetric: state.modelPieMetric,
                apikeyPieMetric: state.apikeyPieMetric,
            }),
        }
    )
);
