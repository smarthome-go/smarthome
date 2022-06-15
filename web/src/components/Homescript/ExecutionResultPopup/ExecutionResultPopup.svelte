<script lang="ts">
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";
    import Button, { Label } from "@smui/button";
    import type {
        homescriptError,
        homescriptResponseWrapper,
    } from "../../../homescript";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    // Keeps track of wether the dialog should be open or not
    export let open: boolean = false;

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

<Dialog
    bind:open
    aria-labelledby="title"
    aria-describedby="content"
    fullscreen
    on:SMUIDialog:closed={() => dispatch("close", null)}
>
    <Header>
        <Title id="title">Result of {data.response.id}</Title>
    </Header>
    <Content id="content">
        <div class="status mdc-elevation-z1">
            <h6>Summary</h6>
            <div class="status__container">
                <div class="status__group">
                    <div
                        class="status__indicator mdc-elevation-z3"
                        class:failure={!data.response.success}
                    >
                        <i class="material-icons"
                            >{data.response.success ? "check" : "error"}</i
                        >
                        {#if data.modeLint}
                            {data.response.success ? "No errors" : "Errors detected"}
                        {:else}
                            {data.response.success ? "Success" : "Failure"}
                        {/if}
                    </div>
                    <div class="status__summary">
                        Exit Code:
                        <code>
                            {data.response.exitCode}
                        </code>
                    </div>
                </div>
                <div class="status__group">
                    {#if !data.response.success}
                        <div class="status__error">
                            <i class="material-icons">
                                {#if data.response.error[0].errorType === "SyntaxError"}
                                    code
                                {:else if data.response.error[0].errorType === "TypeError"}
                                    tag
                                {:else if data.response.error[0].errorType === "ReferenceError"}
                                    alt_route
                                {:else if data.response.error[0].errorType === "ValueError"}
                                    rule
                                {:else if data.response.error[0].errorType === "RuntimeError"}
                                    running_with_errors
                                {:else if data.response.error[0].errorType === "Panic"}
                                    sms_failed
                                {:else}
                                    error
                                {/if}
                            </i>
                            <code>
                                {data.response.error[0].errorType}
                            </code>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
        <div class="output mdc-elevation-z1">
            <h6>Output</h6>
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
                Homescript stopped with exit code
                {data.response.exitCode}
            </span>
        </div>
    </Content>
    <Actions>
        <Button
            on:click={() => {
                dispatch("close", null);
            }}
        >
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .status {
        background-color: var(--clr-height-0-1);
        padding: 1rem 1.5rem;
        border-radius: 0.3rem;

        &__container {
            margin-top: 1rem;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        &__group {
            display: flex;
            align-items: center;
            gap: 1rem;
        }

        // Summarizes the script's outcome
        &__indicator {
            border: 0.01rem solid var(--clr-success);
            color: var(--clr-success);
            border-radius: 0.4rem;
            padding: 0.3rem 0.6rem;
            display: flex;
            align-items: center;
            gap: 0.6rem;

            i {
                opacity: 80%;
                font-size: 1.2rem;
            }

            &.failure {
                border-color: var(--clr-error);
                color: var(--clr-error);
            }
        }

        // Is shown when the script returned an error
        &__error {
            border: 0.01rem solid var(--clr-error);
            color: var(--clr-error);
            border-radius: 0.4rem;
            padding: 0.3rem 0.6rem;
            display: flex;
            align-items: center;
            gap: 0.6rem;

            code {
                color: var(--clr-error);
            }

            i {
                opacity: 80%;
                font-size: 1.2rem;
            }
        }
    }
    .output {
        background-color: var(--clr-height-0-1);
        padding: 1rem 1.5rem;
        margin-top: 1rem;
        border-radius: 0.3rem;
        font-family: "JetBrains Mono", monospace;
        font-size: 0.9rem;
        overflow-wrap: break-word;

        span {
            overflow-wrap: break-word;
        }
    }

    h6 {
        margin: 0;
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
