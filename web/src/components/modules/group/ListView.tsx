'use client';

import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { ChevronDown, Trash2, X, Pencil, Layers } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { type Group, useDeleteGroup, useUpdateGroup, GroupMode, type GroupUpdateRequest } from '@/api/endpoints/group';
import { useModelChannelList } from '@/api/endpoints/model';
import { useTranslations } from 'next-intl';
import { cn } from '@/lib/utils';
import { toast } from '@/components/common/Toast';
import { CopyIconButton } from '@/components/common/CopyButton';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/animate-ui/components/animate/tooltip';
import { getModelIcon } from '@/lib/model-icons';
import { MODE_LABELS, modelChannelKey, buildChannelNameByModelKey } from './utils';
import type { SelectedMember } from './ItemList';
import { GroupEditor, type GroupEditorValues } from './Editor';
import {
    MorphingDialog,
    MorphingDialogClose,
    MorphingDialogContainer,
    MorphingDialogContent,
    MorphingDialogDescription,
    MorphingDialogTitle,
    MorphingDialogTrigger,
    useMorphingDialog,
} from '@/components/ui/morphing-dialog';

interface GroupListViewProps {
    groups: Group[];
}

/* ---------- Edit dialog (reuses GroupEditor) ---------- */

interface EditDialogContentProps {
    group: Group;
    displayMembers: SelectedMember[];
    isSubmitting: boolean;
    onSubmit: (values: GroupEditorValues, onDone?: () => void) => void;
}

function EditDialogContent({ group, displayMembers, isSubmitting, onSubmit }: EditDialogContentProps) {
    const { setIsOpen } = useMorphingDialog();
    const t = useTranslations('group');
    return (
        <>
            <MorphingDialogTitle className="shrink-0">
                <header className="mb-3 flex items-center justify-between">
                    <h2 className="text-2xl font-bold text-card-foreground">
                        {t('detail.actions.edit')}
                    </h2>
                    <MorphingDialogClose className="relative right-0 top-0" />
                </header>
            </MorphingDialogTitle>
            <MorphingDialogDescription className="flex-1 min-h-0 overflow-hidden">
                <GroupEditor
                    key={`edit-group-${group.id}`}
                    initial={{
                        name: group.name,
                        match_regex: group.match_regex ?? '',
                        mode: group.mode,
                        first_token_time_out: group.first_token_time_out ?? 0,
                        session_keep_time: group.session_keep_time ?? 0,
                        members: displayMembers,
                    }}
                    submitText={t('detail.actions.save')}
                    submittingText={t('create.submitting')}
                    isSubmitting={isSubmitting}
                    onCancel={() => setIsOpen(false)}
                    onSubmit={(v) => onSubmit(v, () => setIsOpen(false))}
                />
            </MorphingDialogDescription>
        </>
    );
}

/* ---------- Expanded secondary menu ---------- */

interface ExpandedSecondaryMenuProps {
    group: Group;
}

function ExpandedSecondaryMenu({ group }: ExpandedSecondaryMenuProps) {
    const t = useTranslations('group');
    const { data: modelChannels = [] } = useModelChannelList();

    const channelNameByKey = useMemo(() => {
        const map = new Map<string, string>();
        modelChannels.forEach((mc) => {
            map.set(modelChannelKey(mc.channel_id, mc.name), mc.channel_name);
        });
        return map;
    }, [modelChannels]);

    const enabledByKey = useMemo(() => {
        const map = new Map<string, boolean>();
        modelChannels.forEach((mc) => {
            map.set(modelChannelKey(mc.channel_id, mc.name), mc.enabled);
        });
        return map;
    }, [modelChannels]);

    const displayMembers = useMemo(
        () =>
            [...(group.items || [])]
                .sort((a, b) => a.priority - b.priority)
                .map((item) => ({
                    key: modelChannelKey(item.channel_id, item.model_name),
                    modelName: item.model_name,
                    channelId: item.channel_id,
                    channelName:
                        channelNameByKey.get(modelChannelKey(item.channel_id, item.model_name)) ??
                        `Channel ${item.channel_id}`,
                    weight: item.weight,
                    enabled: enabledByKey.get(modelChannelKey(item.channel_id, item.model_name)) ?? true,
                })),
        [group.items, channelNameByKey, enabledByKey]
    );

    const isEmpty = displayMembers.length === 0;

    return (
        <div className="border-t border-border/40 bg-muted/10">
            <div className="p-4 space-y-4">
                {isEmpty ? (
                    <div className="flex flex-col items-center justify-center gap-2 py-8 text-muted-foreground">
                        <Layers className="size-8 opacity-40" />
                        <span className="text-sm">{t('card.empty')}</span>
                    </div>
                ) : (
                    <div className="grid gap-2">
                        <div className="flex items-center justify-between">
                            <span className="text-xs font-medium text-muted-foreground">
                                {t('form.items')}
                                <span className="ml-1">({displayMembers.length})</span>
                            </span>
                        </div>
                        <div className="grid gap-1.5">
                            {displayMembers.map((member, index) => {
                                const { Avatar } = getModelIcon(member.modelName);
                                const isDisabled = member.enabled === false;
                                return (
                                    <div
                                        key={member.key}
                                        className={cn(
                                            'flex items-center gap-3 rounded-lg border border-border/40 bg-background px-3 py-2.5 transition-colors hover:bg-muted/40',
                                            isDisabled && 'opacity-60 grayscale'
                                        )}
                                    >
                                        <span className="flex size-5 shrink-0 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary">
                                            {index + 1}
                                        </span>
                                        <span className="flex shrink-0 items-center">
                                            <Avatar size={16} />
                                        </span>
                                        <div className="flex min-w-0 flex-1 flex-col">
                                            <Tooltip side="top" sideOffset={10} align="start">
                                                <TooltipTrigger className={cn('text-sm font-medium truncate leading-tight', isDisabled && 'text-muted-foreground')}>
                                                    {member.modelName}
                                                </TooltipTrigger>
                                                <TooltipContent key={member.modelName}>{member.modelName}</TooltipContent>
                                            </Tooltip>
                                            <span className="text-[11px] text-muted-foreground truncate leading-tight">
                                                {member.channelName}
                                            </span>
                                        </div>
                                        {group.mode === GroupMode.Weighted && (
                                            <span className="shrink-0 text-xs font-medium text-muted-foreground">
                                                w={member.weight ?? 1}
                                            </span>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}

                <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground pt-1 border-t border-border/20">
                    <span>
                        {t('mode.' + MODE_LABELS[group.mode])}
                    </span>
                    {group.match_regex && (
                        <span className="truncate max-w-[200px]" title={group.match_regex}>
                            regex: {group.match_regex}
                        </span>
                    )}
                    {group.first_token_time_out !== undefined && group.first_token_time_out > 0 && (
                        <span>
                            timeout: {group.first_token_time_out}s
                        </span>
                    )}
                    {group.session_keep_time !== undefined && group.session_keep_time > 0 && (
                        <span>
                            session: {group.session_keep_time}s
                        </span>
                    )}
                </div>
            </div>
        </div>
    );
}

/* ---------- Individual list row ---------- */

function GroupListRow({ group }: { group: Group }) {
    const t = useTranslations('group');
    const updateGroup = useUpdateGroup();
    const deleteGroup = useDeleteGroup();
    const { data: modelChannels = [] } = useModelChannelList();

    const [expanded, setExpanded] = useState(false);
    const [confirmDelete, setConfirmDelete] = useState(false);

    /* ---- member / editing state ---- */
    const [members, setMembers] = useState<SelectedMember[]>([]);
    const isDragging = useRef(false);
    const weightTimerRef = useRef<NodeJS.Timeout | null>(null);
    const membersRef = useRef<SelectedMember[]>([]);

    const channelNameByKey = useMemo(() => buildChannelNameByModelKey(modelChannels), [modelChannels]);
    const enabledByKey = useMemo(() => {
        const map = new Map<string, boolean>();
        modelChannels.forEach((mc) => {
            map.set(modelChannelKey(mc.channel_id, mc.name), mc.enabled);
        });
        return map;
    }, [modelChannels]);

    const displayMembers = useMemo((): SelectedMember[] =>
        [...(group.items || [])]
            .sort((a, b) => a.priority - b.priority)
            .map((item) => ({
                id: modelChannelKey(item.channel_id, item.model_name),
                name: item.model_name,
                enabled: enabledByKey.get(modelChannelKey(item.channel_id, item.model_name)) ?? true,
                channel_id: item.channel_id,
                channel_name: channelNameByKey.get(modelChannelKey(item.channel_id, item.model_name)) ?? `Channel ${item.channel_id}`,
                item_id: item.id,
                weight: item.weight,
            })),
        [group.items, channelNameByKey, enabledByKey]
    );

    useEffect(() => {
        if (!isDragging.current) setMembers([...displayMembers]);
    }, [displayMembers]);

    useEffect(() => {
        membersRef.current = members;
    }, [members]);

    useEffect(() => {
        return () => { if (weightTimerRef.current) clearTimeout(weightTimerRef.current); };
    }, []);

    const isUpdatingMode = (() => {
        if (!updateGroup.isPending) return false;
        const v = updateGroup.variables;
        if (typeof v !== 'object' || v === null) return false;
        return 'mode' in v && typeof (v as { mode?: unknown }).mode === 'number';
    })();

    const priorityByItemId = useMemo(() => {
        const map = new Map<number, number>();
        (group.items || []).forEach((item) => {
            if (item.id !== undefined) map.set(item.id, item.priority);
        });
        return map;
    }, [group.items]);

    const memberCount = group.items?.length ?? 0;

    const onSuccess = useCallback(() => toast.success(t('toast.updated')), [t]);
    const onError = useCallback((error: Error) => toast.error(t('toast.updateFailed'), { description: error.message }), [t]);

    const handleSubmitEdit = useCallback((values: GroupEditorValues, onDone?: () => void) => {
        if (!group.id) return;

        const originalItems = [...(group.items || [])].sort((a, b) => a.priority - b.priority);
        const originalById = new Map<number, { priority: number; weight: number }>();
        const originalIds = new Set<number>();
        originalItems.forEach((it) => {
            if (typeof it.id === 'number') {
                originalIds.add(it.id);
                originalById.set(it.id, { priority: it.priority, weight: it.weight });
            }
        });

        const newIds = new Set<number>();
        values.members.forEach((m) => { if (typeof m.item_id === 'number') newIds.add(m.item_id); });

        const items_to_delete = Array.from(originalIds).filter((id) => !newIds.has(id));

        const items_to_add = values.members
            .map((m, idx) => ({ m, priority: idx + 1 }))
            .filter(({ m }) => typeof m.item_id !== 'number')
            .map(({ m, priority }) => ({
                channel_id: m.channel_id,
                model_name: m.name,
                priority,
                weight: m.weight ?? 1,
            }));

        const items_to_update = values.members
            .map((m, idx) => ({ m, priority: idx + 1 }))
            .filter(({ m }) => typeof m.item_id === 'number')
            .map(({ m, priority }) => {
                const id = m.item_id!;
                const orig = originalById.get(id);
                const weight = m.weight ?? 1;
                if (!orig) return null;
                if (orig.priority === priority && orig.weight === weight) return null;
                return { id, priority, weight };
            })
            .filter((x): x is { id: number; priority: number; weight: number } => x !== null);

        const payload: GroupUpdateRequest = { id: group.id };
        const nextName = values.name.trim();
        const nextRegex = (values.match_regex ?? '').trim();
        const nextFirstTokenTimeOut = values.first_token_time_out ?? 0;
        const nextSessionKeepTime = values.session_keep_time ?? 0;

        if (nextName && nextName !== group.name) payload.name = nextName;
        if (values.mode !== group.mode) payload.mode = values.mode;
        if (nextRegex !== (group.match_regex ?? '')) payload.match_regex = nextRegex;
        if (nextFirstTokenTimeOut !== (group.first_token_time_out ?? 0)) payload.first_token_time_out = nextFirstTokenTimeOut;
        if (nextSessionKeepTime !== (group.session_keep_time ?? 0)) payload.session_keep_time = nextSessionKeepTime;
        if (items_to_add.length) payload.items_to_add = items_to_add;
        if (items_to_update.length) payload.items_to_update = items_to_update;
        if (items_to_delete.length) payload.items_to_delete = items_to_delete;

        if (Object.keys(payload).length === 1) {
            onDone?.();
            return;
        }

        updateGroup.mutate(payload, {
            onSuccess: () => {
                onSuccess();
                onDone?.();
            },
            onError,
        });
    }, [group, onSuccess, onError, updateGroup]);

    return (
        <div
            className={cn(
                'rounded-2xl border border-border bg-card text-card-foreground transition-all duration-200',
                expanded ? 'shadow-md' : 'custom-shadow hover:shadow-md',
            )}
        >
            {/* ---- Header row ---- */}
            <div
                className="flex items-center gap-3 px-4 py-3 cursor-pointer select-none transition-colors hover:bg-muted/30 rounded-t-2xl"
                onClick={() => setExpanded((prev) => !prev)}
            >
                <motion.div
                    animate={{ rotate: expanded ? 0 : -90 }}
                    transition={{ duration: 0.2, ease: 'easeInOut' }}
                    className="flex size-6 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground"
                >
                    <ChevronDown className="size-3.5" />
                </motion.div>

                <span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary">
                    {memberCount}
                </span>

                <div className="flex min-w-0 flex-1 items-center gap-2">
                    <Tooltip side="top" sideOffset={10} align="start">
                        <TooltipTrigger asChild>
                            <h3 className="text-sm font-semibold truncate">{group.name}</h3>
                        </TooltipTrigger>
                        <TooltipContent key={group.name}>{group.name}</TooltipContent>
                    </Tooltip>
                    <span
                        className={cn(
                            'shrink-0 rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-tight',
                            group.mode === GroupMode.RoundRobin && 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
                            group.mode === GroupMode.Random && 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
                            group.mode === GroupMode.Failover && 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
                            group.mode === GroupMode.Weighted && 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
                        )}
                    >
                        {t('mode.' + MODE_LABELS[group.mode])}
                    </span>
                </div>

                {/* ---- Actions (stop propagation so clicks don't toggle) ---- */}
                <div className="flex items-center gap-1 shrink-0" onClick={(e) => e.stopPropagation()}>
                    {/* Edit */}
                    <MorphingDialog>
                        <MorphingDialogTrigger className="p-1.5 rounded-lg transition-colors hover:bg-muted text-muted-foreground hover:text-foreground">
                            <Tooltip side="top" sideOffset={10} align="center">
                                <TooltipTrigger asChild>
                                    <Pencil className="size-3.5" />
                                </TooltipTrigger>
                                <TooltipContent>{t('detail.actions.edit')}</TooltipContent>
                            </Tooltip>
                        </MorphingDialogTrigger>

                        <MorphingDialogContainer>
                            <MorphingDialogContent className="relative w-screen max-w-full md:max-w-4xl bg-card text-card-foreground px-6 py-4 rounded-3xl h-[calc(100vh-2rem)] flex flex-col overflow-hidden">
                                <EditDialogContent
                                    group={group}
                                    displayMembers={displayMembers}
                                    isSubmitting={updateGroup.isPending}
                                    onSubmit={handleSubmitEdit}
                                />
                            </MorphingDialogContent>
                        </MorphingDialogContainer>
                    </MorphingDialog>

                    {/* Copy name */}
                    <Tooltip side="top" sideOffset={10} align="center">
                        <TooltipTrigger asChild>
                            <CopyIconButton
                                text={group.name}
                                className="p-1.5 rounded-lg transition-colors hover:bg-muted text-muted-foreground hover:text-foreground"
                                copyIconClassName="size-3.5"
                                checkIconClassName="size-3.5 text-primary"
                            />
                        </TooltipTrigger>
                        <TooltipContent>{t('detail.actions.copyName')}</TooltipContent>
                    </Tooltip>

                    {/* Delete */}
                    {!confirmDelete && (
                        <Tooltip side="top" sideOffset={10} align="center">
                            <TooltipTrigger asChild>
                                <motion.button
                                    layoutId={`list-delete-btn-${group.id}`}
                                    type="button"
                                    onClick={() => setConfirmDelete(true)}
                                    className="p-1.5 rounded-lg hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                                >
                                    <Trash2 className="size-3.5" />
                                </motion.button>
                            </TooltipTrigger>
                            <TooltipContent>{t('detail.actions.delete')}</TooltipContent>
                        </Tooltip>
                    )}

                    <AnimatePresence>
                        {confirmDelete && (
                            <motion.div
                                layoutId={`list-delete-btn-${group.id}`}
                                className="flex items-center gap-1.5 bg-destructive p-1 rounded-lg"
                                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                            >
                                <button
                                    type="button"
                                    onClick={() => setConfirmDelete(false)}
                                    className="flex size-6 items-center justify-center rounded-md bg-destructive-foreground/20 text-destructive-foreground transition-all hover:bg-destructive-foreground/30 active:scale-95"
                                >
                                    <X className="size-3" />
                                </button>
                                <button
                                    type="button"
                                    onClick={() =>
                                        group.id &&
                                        deleteGroup.mutate(group.id, {
                                            onSuccess: () => toast.success(t('toast.deleted')),
                                        })
                                    }
                                    disabled={deleteGroup.isPending}
                                    className="flex h-6 items-center justify-center gap-1 rounded-md bg-destructive-foreground text-destructive px-2 text-xs font-semibold transition-all hover:bg-destructive-foreground/90 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    <Trash2 className="size-3" />
                                    {t('detail.actions.confirmDelete')}
                                </button>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
            </div>

            {/* ---- Expanded area ---- */}
            <AnimatePresence initial={false}>
                {expanded && (
                    <motion.div
                        key="secondary-menu"
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        transition={{ duration: 0.25, ease: 'easeInOut' }}
                        className="overflow-hidden"
                    >
                        {/* Mode quick-switch */}
                        <div className="px-4 pt-3 pb-1">
                            <div className="flex gap-1">
                                {([GroupMode.RoundRobin, GroupMode.Random, GroupMode.Failover, GroupMode.Weighted] as const).map((m) => (
                                    <button
                                        key={m}
                                        type="button"
                                        aria-disabled={isUpdatingMode || !group.id}
                                        onClick={() => {
                                            if (isUpdatingMode || !group.id) return;
                                            if (m === group.mode) return;
                                            updateGroup.mutate({ id: group.id!, mode: m }, { onSuccess, onError });
                                        }}
                                        className={cn(
                                            'flex-1 py-1 text-xs rounded-lg transition-colors',
                                            group.mode === m ? 'bg-primary text-primary-foreground' : 'bg-muted hover:bg-muted/80',
                                            (!group.id) && 'cursor-not-allowed opacity-50'
                                        )}
                                    >
                                        {t(`mode.${MODE_LABELS[m]}`)}
                                    </button>
                                ))}
                            </div>
                        </div>

                        <ExpandedSecondaryMenu group={group} />
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}

/* ---------- List view container ---------- */

export function GroupListView({ groups }: GroupListViewProps) {
    const t = useTranslations('group');

    if (groups.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center gap-3 py-20 text-muted-foreground">
                <Layers className="size-12 opacity-30" />
                <p className="text-sm">{t('empty')}</p>
            </div>
        );
    }

    return (
        <div className="h-full min-h-0 overflow-y-auto">
            <div className="flex flex-col gap-2 pb-4">
                {groups.map((group, index) => (
                    <GroupListRow key={group.id ?? `group-${index}`} group={group} />
                ))}
            </div>
        </div>
    );
}
