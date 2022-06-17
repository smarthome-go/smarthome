<script lang="ts">
    import type {
        homescriptError,
        homescriptResponseWrapper,
    } from "../../../homescript";

    // Data is bound to display the result
    export let data: homescriptResponseWrapper;

    function errToHtml(err: homescriptError, programCode: string): string {
        const lines = programCode.split("\n");
        let line1 = "";

        if (err.location.line > 1)
            line1 = `<br>&nbsp;<span class="gray">${(err.location.line - 1)
                .toString()
                .padStart(3, " ")
                .replaceAll(" ", "&nbsp;")}&nbsp;|&nbsp;</span>${
                lines[err.location.line - 2]
            }`;

        const line2 = `&nbsp;<span class="gray">${(err.location.line - 0)
            .toString()
            .padStart(3, " ")
            .replaceAll(" ", "&nbsp;")}&nbsp;|&nbsp;</span>${
            lines[err.location.line - 1]
        }`;

        let line3 = "";
        if (err.location.line > lines.length)
            line1 = `<br>&nbsp;<span class="gray">${(err.location.line + 1)
                .toString()
                .padStart(3, " ")
                .replaceAll(" ", "&nbsp;")}&nbsp;|&nbsp;</span>${
                lines[err.location.line]
            }`;

        const marker = `${"&nbsp;".repeat(
            err.location.column + 6
        )}<span class="red bold">^</span>`;

        return (
            `<span class="cyan bold">${err.errorType}</span><span class="bold">&nbsp;at&nbsp;${err.location.filename}:${err.location.line}:${err.location.column}</span>` +
            `<br>${line1}<br>${line2}<br>${marker}${line3}<br><br><span class="red bold">${err.message
                .replaceAll(" ", "&nbsp;")
                .replaceAll("\n", "<br>")}</span>`
        );
    }
</script>

<div class="terminal">
    {#if data.response.output.length > 0}
        {@html data.response.output
            .replaceAll("\n", "<br>")
            .replaceAll(" ", "&nbsp;")}
        <br />
    {/if}
    {#if !data.response.success}
        <br />
        {#each data.response.error as err}
            {@html errToHtml(err, data.code)}
        {/each}
        <br />
        <br />
    {/if}
    <span class="text-disabled">
        {#if data.modeLint}
            Homescript stopped with exit code
            {data.response.exitCode}
        {:else}
            Lint finished with exit code:
            {data.response.exitCode}
        {/if}
    </span>
</div>

<style lang="scss">
    .terminal {
        font-family: "JetBrains Mono", monospace;
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
            color: #4f5666;
        }
    }
</style>
