<script lang="ts">
    import IconButton from '@smui/icon-button'
    import { createSnackbar, data } from '../global'
    import Progress from './Progress.svelte'
    import { marked } from 'marked'

    export let dummy = false

    export let id = 0
    export let priority = 0
    export let name = ''
    export let description = ''
    export let date = 0

    let loading = false
    let deleted = false
    let container: HTMLDivElement
    const priorityColors = ['var(--clr-success)', 'var(--clr-warn)', 'var(--clr-error)']

    $: if (deleted) {
        container.style.setProperty('--height', container.getBoundingClientRect().height + 'px')
        container.getBoundingClientRect()
        container.style.height = '0'
    }

    async function deleteSelf() {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/notification/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error()
            deleted = true
            setTimeout(() => {
                $data.notifications = $data.notifications.filter(n => n.id !== id)
            }, 300)
        } catch {
            $createSnackbar('Could not delete notification')
        }
        loading = false
    }

    const renderer = {
        code(code: string, infostring: string, escaped: boolean) {
            let outputCode = code
                .replaceAll('\n', '<br/>')
                .replaceAll(' ', '&nbsp;<wbr>')
                .replaceAll('\t', '&nbsp;&nbsp;&nbsp;&nbsp;<wbr>')

            return `<code class='md-code'>${outputCode}</code>`
        },
        codespan(code: string) {
            let outputCode = code
                .replaceAll(' ', '&nbsp;<wbr>')
                .replaceAll('\t', '&nbsp;&nbsp;&nbsp;&nbsp;<wbr>')

            return `<code class='md-codespan'>${outputCode}</code>`
        },
    }

    marked.use({ renderer })
    let html = marked.parse(description)
</script>

<div
    class="root mdc-elevation--z2"
    bind:this={container}
    class:deleted
    class:dummy
    id={`notification-${id}`}
>
    {#if dummy}
        <div class="title" />
        <div class="description" />
        <div class="description small" />
        <div class="time" />
    {:else}
        <div class="line" style:--clr-priority={priorityColors[priority - 1]} />
        <Progress class="spinner" bind:loading type="circular" />
        <IconButton class="delete material-icons" title="Delete" on:click={deleteSelf}
            >delete</IconButton
        >
        <h6>{name}</h6>
        <p>
            {@html html}
        </p>
        <p class="date text-hint">
            {new Date(
                date - 2 * 60 * 60 * 1000 /* Subtract two hours off the time */,
            ).toLocaleString()}
        </p>
    {/if}
</div>

<style lang="scss">
    .root {
        border-radius: 0.3rem;
        padding: 1rem;
        padding-left: 1.5rem;
        white-space: normal;
        position: relative;
        transition-property: transform, height, margin-bottom, padding;
        transition-duration: 0.3s;
        width: 100%;
        box-sizing: border-box;
        user-select: text;
        background-color: var(--clr-height-8-2);
        overflow: hidden;
        height: var(--height);
        margin-bottom: 1rem;
        flex-shrink: 0;

        h6 {
            margin: 0;
        }

        .date {
            font-size: 0.7rem;
        }

        &.deleted {
            transform: translateX(-110%);
            margin-bottom: 0;
            padding: 0 1rem;
        }

        &.dummy {
            animation-iteration-count: infinite;
            animation-duration: 2s;
            animation-name: loading;

            .title {
                background-color: var(--clr-height-0-8);
                height: 1.4rem;
                width: 5rem;
                margin-bottom: 1rem;
            }

            .description {
                background-color: var(--clr-height-0-16);
                height: 1rem;
                width: 100%;
                margin-bottom: 0.5rem;
                &.small {
                    width: 90%;
                }
            }

            .time {
                background-color: var(--clr-height-0-12);
                height: 0.6rem;
                width: 6rem;
                margin-top: 1rem;
            }
        }
    }

    @keyframes loading {
        0% {
            filter: brightness(100%);
        }
        50% {
            filter: brightness(90%);
        }
        100% {
            filter: brightness(100%);
        }
    }

    .line {
        position: absolute;
        border-radius: 0.3rem;
        overflow: hidden;
        top: 0;
        left: 0;
        bottom: 0;
        width: 0.5rem;

        &::before {
            content: '';
            position: absolute;
            top: 0;
            bottom: 0;
            left: 0;
            width: 0.3rem;
            background-color: var(--clr-priority);
        }
    }

    .root :global .delete {
        position: absolute;
        top: 0.25rem;
        right: 0.25rem;
        cursor: pointer;
        z-index: 10;
    }

    .root > :global .spinner {
        position: absolute;
        top: 1rem;
        right: 4rem;
        opacity: 0;
        &.visible {
            opacity: 1;
        }
    }

    // Markdown styles
    :global .md-code,
    :global .md-codespan {
        background-color: var(--clr-height-2-1);
        box-sizing: border-box;
        border-radius: 0.25rem;
    }

    :global .md-codespan {
        padding: 0.1rem 0.2rem;
    }

    :global .md-code {
        padding: 0.2rem 0.5rem;
        display: block;
        width: 100%;
    }

    :global table,
    :global td,
    :global th {
        border: 1px solid var(--clr-text-disabled);
    }

    :global td {
        padding: 0.2rem 0.5rem;
    }

    :global table {
        width: 100%;
        border-collapse: collapse;
    }
</style>
