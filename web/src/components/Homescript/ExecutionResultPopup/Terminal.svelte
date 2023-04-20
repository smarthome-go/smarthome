<script lang="ts">
    import type { hmsResWrapper } from 'src/pages/hmsEditor/websocket'

    import type { homescriptError } from '../../../homescript'

    // Data is bound to display the result
    export let data: hmsResWrapper

    // Terminal output
    export let output: string

    function errToHtml(err: homescriptError, data: hmsResWrapper): string {
        let code = data.code

        if (data.fileContents[err.span.filename] !== undefined) {
            code = data.fileContents[err.span.filename]
        }

        const lines = code.split('\n')

        let line1 = ''
        if (err.span.start.line > 1)
            line1 = `<br>&nbsp;<span class="gray">${(err.span.start.line - 1)
                .toString()
                .padStart(3, ' ')
                .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line - 2]
                .replaceAll('\t', '    ')
                .replaceAll(' ', '&nbsp;')}`

        const line2 = `&nbsp;<span class="gray">${err.span.start.line
            .toString()
            .padStart(3, ' ')
            .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line - 1]
            .replaceAll('\t', '    ')
            .replaceAll(' ', '&nbsp;')}`

        let line3 = ''
        if (err.span.start.line < lines.length)
            line3 = `<br>&nbsp;<span class="gray">${(err.span.start.line + 1)
                .toString()
                .padStart(3, ' ')
                .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line]
                .replaceAll('\t', '    ')
                .replaceAll(' ', '&nbsp;')}`

        let color = 'red'
        if (err.kind == 'Warning') {
            color = 'yellow'
        }
        if (err.kind == 'Info') {
            color = 'cyan'
        }

        let rawMarker = '^'
        if (err.kind === 'Warning' || err.kind === 'Info') {
            rawMarker = '~'
        }
        if (err.span.start.line === err.span.end.line) {
            rawMarker = rawMarker.repeat(err.span.end.column - err.span.start.column + 1)
        } else {
            rawMarker = '^'
        }

        const marker = `${'&nbsp;'.repeat(
            err.span.start.column + 6,
        )}<span class="${color} bold">${rawMarker}</span>`

        return (
            `<span class="${color} bold">${err.kind}</span><span class="bold">&nbsp;at&nbsp;${err.span.filename}:${err.span.start.line}:${err.span.start.column}</span>` +
            `<br>${line1}<br>${line2}<br>${marker}${line3}<br><br><span class="${color} bold">${err.message
                .replaceAll(' ', '&nbsp;')
                .replaceAll('\n', '<br>')}</span>`
        )
    }
</script>

<div class="terminal">
    {#if output.length > 0}
        {@html output.replaceAll('\n', '<br>').replaceAll(' ', '&nbsp;')}
        <br />
    {/if}
    {#if data !== undefined}
        {#if data.exitCode !== 0 && output.length > 0}
            <br />
        {/if}
        {#each data.errors as err}
            {@html errToHtml(err, data)}
            <br />
            <br />
        {/each}
        <span class="text-disabled">
            {#if data.modeRun}
                Homescript stopped with exit code
                {data.exitCode}
            {:else}
                Lint finished with exit code
                {data.exitCode}
            {/if}
        </span>
    {:else}
        <br />
    {/if}
</div>

<style lang="scss">
    .terminal {
        font-family: 'JetBrains Mono', monospace;
        font-size: 0.9rem;
        overflow-wrap: break-word;
    }

    :global {
        .bold {
            font-weight: bold;
        }

        .red {
            color: #ff616e;
        }

        .yellow {
            color: #ffef55;
        }

        .cyan {
            color: #4cd1e0;
        }

        .gray {
            color: #7d7e82;
        }
    }
</style>
