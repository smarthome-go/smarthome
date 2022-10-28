<script lang="ts">
    import type { homescriptError, homescriptResponseWrapper } from '../../../homescript'

    // Data is bound to display the result
    export let data: homescriptResponseWrapper

    function errToHtml(err: homescriptError, programCode: string): string {
        const lines = programCode.split('\n')
        let line1 = ''

        if (err.span.start.line > 1)
            line1 = `<br>&nbsp;<span class="gray">${(err.span.start.line - 1)
                .toString()
                .padStart(3, ' ')
                .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line - 2]}`

        const line2 = `&nbsp;<span class="gray">${(err.span.start.line - 0)
            .toString()
            .padStart(3, ' ')
            .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line - 1]}`

        let line3 = ''
        if (err.span.start.line > lines.length)
            line1 = `<br>&nbsp;<span class="gray">${(err.span.start.line + 1)
                .toString()
                .padStart(3, ' ')
                .replaceAll(' ', '&nbsp;')}&nbsp;|&nbsp;</span>${lines[err.span.start.line]}`

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
        }

        const marker = `${'&nbsp;'.repeat(
            err.span.start.column + 6,
        )}<span class="${color} bold">${rawMarker}</span>`

        return (
            `<span class="${color} bold">${
                err.kind
            }</span><span class="bold">&nbsp;at&nbsp;${'SOME-FILE'}:${err.span.start.line}:${
                err.span.start.column
            }</span>` +
            `<br>${line1}<br>${line2}<br>${marker}${line3}<br><br><span class="${color} bold">${err.message
                .replaceAll(' ', '&nbsp;')
                .replaceAll('\n', '<br>')}</span>`
        )
    }
</script>

<div class="terminal">
    {#if data.response.output.length > 0}
        {@html data.response.output.replaceAll('\n', '<br>').replaceAll(' ', '&nbsp;')}
        <br />
    {/if}
    {#if !data.response.success && data.response.output.length > 0}
        <br />
    {/if}
    {#each data.response.errors as err}
        {@html errToHtml(err, data.code)}
        <br />
        <br />
    {/each}
    <span class="text-disabled">
        {#if data.modeRun}
            Homescript stopped with exit code
            {data.response.exitCode}
        {:else}
            Lint finished with exit code
            {data.response.exitCode}
        {/if}
    </span>
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

        .cyan {
            color: #4cd1e0;
        }

        .gray {
            color: #7d7e82;
        }
    }
</style>
