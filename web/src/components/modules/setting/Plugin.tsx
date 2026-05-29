'use client';

import { useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { BrainCircuit, HelpCircle } from 'lucide-react';
import { Switch } from '@/components/ui/switch';
import { useSettingList, useSetSetting, SettingKey } from '@/api/endpoints/setting';
import { toast } from '@/components/common/Toast';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/animate-ui/components/animate/tooltip';

export function SettingPlugin() {
    const t = useTranslations('setting');
    const { data: settings } = useSettingList();
    const setSetting = useSetSetting();

    const [thinkingEnabled, setThinkingEnabled] = useState(true);
    const initialThinkingEnabled = useRef(true);

    useEffect(() => {
        if (!settings) return;

        const setting = settings.find(s => s.key === SettingKey.EnableThinkingPlugin);
        if (!setting) return;

        const enabled = setting.value === 'true';
        queueMicrotask(() => setThinkingEnabled(enabled));
        initialThinkingEnabled.current = enabled;
    }, [settings]);

    const handleThinkingEnabledChange = (checked: boolean) => {
        setThinkingEnabled(checked);
        setSetting.mutate(
            { key: SettingKey.EnableThinkingPlugin, value: checked ? 'true' : 'false' },
            {
                onSuccess: () => {
                    toast.success(t('saved'));
                    initialThinkingEnabled.current = checked;
                },
                onError: () => {
                    setThinkingEnabled(initialThinkingEnabled.current);
                }
            }
        );
    };

    return (
        <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
            <h2 className="text-lg font-bold text-card-foreground flex items-center gap-2">
                <BrainCircuit className="h-5 w-5" />
                {t('plugin.title')}
            </h2>

            <div className="flex items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                    <BrainCircuit className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">{t('plugin.thinking.label')}</span>
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <HelpCircle className="size-4 text-muted-foreground cursor-help" />
                            </TooltipTrigger>
                            <TooltipContent>
                                {t('plugin.thinking.hint')}
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
                <Switch
                    checked={thinkingEnabled}
                    onCheckedChange={handleThinkingEnabledChange}
                    disabled={setSetting.isPending}
                />
            </div>
        </div>
    );
}
