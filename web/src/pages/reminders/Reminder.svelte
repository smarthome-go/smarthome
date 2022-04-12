<script lang="ts">
    import IconButton from '@smui/icon-button'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Edit from './Edit.svelte'
    import { reminders } from './main'

    export let id: number
    export let name: string
    export let description: string
    export let priority: number
    export let dueDate: number
    export let createdDate: number
    export let userWasNotified: boolean

    let deleteLoading = false
    let deleted = false

    const priorities = [
        { label: 'LOW', color: 'var(--clr-priority-low)' },
        { label: 'NORMAL', color: 'var(--clr-success)' },
        { label: 'MEDIUM', color: 'var(--clr-priority-medium)' },
        { label: 'HIGH', color: 'var(--clr-warn)' },
        { label: 'URGENT', color: 'var(--clr-error)' },
    ]
    const priorityColor = priorities[priority].color

    function millisToDate(millis: number): string {
        const d = new Date(millis)
        return d.getDate() + '.' + (d.getMonth() + 1) + '.' + d.getFullYear()
    }

    let container: HTMLDivElement
    $: if (deleted) {
        container.style.setProperty(
            '--height',
            container.getBoundingClientRect().height + 'px'
        )
        container.getBoundingClientRect()
        container.style.height = '0'
    }

    async function deleteSelf() {
        deleteLoading = true
        try {
            const res = await (
                await fetch('/api/reminder/delete', {
                    headers: { 'Content-Type': 'application/json' },
                    method: 'DELETE',
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error()
            deleted = true
            setTimeout(() => {
                $reminders = $reminders.filter((n) => n.id !== id)
            }, 300)
        } catch (err) {
            $createSnackbar('Could not mark reminder as completed')
        }
        deleteLoading = false
    }

    async function modify(name, description, priority, dueDate) {
        console.log(id)
    }
</script>

<div
    bind:this={container}
    class="root mdc-elevation--z3"
    class:deleted
    style:--clr-priority={priorityColor}
>
    <div id="top">
        <div class="align">
            <h6>{name}</h6>
            {#if userWasNotified}
                <i title="A notifification was sent" class="material-icons"
                    >notifications_active</i
                >
            {/if}
        </div>
        <div class="align">
            <Progress class="spinner" bind:loading={deleteLoading} type="circular" />
            <p
                style:--clr-priority={priorityColor}
                class="text-hint"
                id="priority"
            >
                {priorities[priority].label}
            </p>
            <IconButton
                title="Mark done"
                class="material-icons"
                on:click={() => deleteSelf()}>done</IconButton
            >
        </div>
    </div>
    <p>{description}</p>
    <div id="bottom">
        <p class="date text-hint">
            Due Date {millisToDate(dueDate)}
        </p>
        <p class="text-disabled date">created {millisToDate(createdDate)}</p>
    </div>
    <Edit modify={modify}/>
</div>

<style lang="scss">
    @use '../../mixins' as *;

    .root {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        border-left: 0.3rem solid var(--clr-priority);
        padding: 1rem;
        transition-property: transform, height, margin-bottom, padding, opacity;
        transition-duration: 0.3s;
        margin-bottom: 1rem;
        word-wrap: break-word;

        &.deleted {
            transform: translateX(-110%);
            margin-bottom: 0;
            padding: 0 1rem;
        }
    }
    h6 {
        margin: 0;
    }
    p {
        margin: 0.5rem 0;
    }
    #top {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    #bottom {
        display: flex;
        justify-content: space-between;
        @include mobile {
            flex-direction: column;
            p {
                margin: 0.1rem;
            }
        }
    }
    .align {
        display: flex;
        align-items: center;
        gap: 0.8rem;
        @include mobile {
            gap: 0.5rem;
        }
    }
    #priority {
        opacity: 80%;
        color: var(--clr-priority);
        @include mobile {
            display: none;
        }
    }
    .date {
        font-size: 0.8rem;
    }
</style>
