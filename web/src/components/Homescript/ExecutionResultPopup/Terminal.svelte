<script lang="ts">
    import type { hmsResWrapper } from 'src/pages/hmsEditor/websocket'

    import type { homescriptError } from '../../../homescript'

    // Data is bound to display the result
    export let data: hmsResWrapper

    // Terminal output
    export let output: string

    function replaceWithHTMLCharacterCodes(input: string): string {
        return input
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll(' ', '&nbsp;')
            .replaceAll('\t', '&#09;')
            .replaceAll('\n', '<br>')
    }

    function errToHtml(err: homescriptError, data: hmsResWrapper): string {
        const fromContents = data.fileContents.get(err.span.filename)

        let code = fromContents
        let missingSources = false

        if (fromContents === undefined || fromContents === null) {
            // throw("Missing Homescript sources for this error")
            console.log("No sources")
            missingSources = true
        }

        let color = 'red'
        let kind = 'error: unknown'
        let message = 'error: unknown'
        let notes = ''

        if (err.syntaxError !== null) {
            kind = 'SyntaxError'
            message = err.syntaxError.message
        } else if (err.diagnosticError !== null) {
            message = err.diagnosticError.message
            notes = err.diagnosticError.notes
                .map(n => `<span class='cyan bold'>- note:</span> ${n}`)
                .join('')

            switch (err.diagnosticError.kind) {
                case 0:
                    kind = 'Hint'
                    color = 'purple'
                    break
                case 1:
                    kind = 'Info'
                    color = 'cyan'
                    break
                case 2:
                    kind = 'Warning'
                    color = 'yellow'
                    break
                case 3:
                    kind = 'Error'
                    color = 'red'
                    break
            }
        } else {
            kind = err.runtimeError.kind
            message = err.runtimeError.message
        }

        // if there is no useful span, do not try to include it
        if (
            err.span.start.line === 0 &&
            err.span.start.column === 0 &&
            err.span.end.line === 0 &&
            err.span.end.column == 0
        ) {
            return `<span class="${color}">${kind}</span> in ${err.span.filename}<br>${message}<br>${notes}`
        }

        if (missingSources) {
            return (
                `<span class="${color} bold">${kind}</span><span class="bold">&nbsp;at&nbsp;${err.span.filename}:${err.span.start.line}:${err.span.start.column}</span>` +
                `<br><br><span class="${color} bold">${message
                    .replaceAll(' ', '&nbsp;')
                    .replaceAll('\n', '<br>')}</span><br>${notes}`
            )
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

        let rawMarker = '^'
        if (color == 'yellow' || color === 'cyan' || color === 'purple') {
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
            `<span class="${color} bold">${kind}</span><span class="bold">&nbsp;at&nbsp;${err.span.filename}:${err.span.start.line}:${err.span.start.column}</span>` +
            `<br>${line1}<br>${line2}<br>${marker}${line3}<br><br><span class="${color} bold">${message
                .replaceAll(' ', '&nbsp;')
                .replaceAll('\n', '<br>')}</span><br>${notes}`
        )
    }
</script>

<div class="terminal">
    {#if output.length > 0}
        {@html replaceWithHTMLCharacterCodes(output)}
        <br />
    {/if}
    {#if data !== undefined}
        {#if !data.success}
            <br />
        {/if}
        {#each data.errors as err}
            {@html errToHtml(err, data)}
            <br />
            <br />
        {/each}
        <span class="text-disabled">
            {#if data.modeRun}
                {#if data.success}
                    Homescript executed successfully
                {:else}
                    Homescript crashed during runtime
                {/if}
            {:else if data.success}
                Analyzer detected no issues
            {:else}
                Analyzer detected issues
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
            color: #4ad0df;
        }

        .purple {
            color: #d472fe;
        }

        .gray {
            color: #7d7e82;
        }
    }
</style>
