<script lang="ts">
    import { createSnackbar, data } from "../../global";
    import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
    import IconButton from "@smui/icon-button";
    import { onMount } from "svelte";
    import Button from "@smui/button";
    import Progress from "../../components/Progress.svelte";
    import DeleteToken from "./dialogs/DeleteToken.svelte";
    import AddToken from "./dialogs/AddToken.svelte";

    let loading = false;

    let addTokenOpen = false;

    // User tokens
    interface UserToken {
        user: string;
        token: string;
        data: {
            label: string;
        };
    }

    let tokens: UserToken[] = [];
    let visibleTokens: string[] = [];
    let tokensForDeletion: string[] = [];

    async function fetchUserTokens() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/user/token/list/personal")
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            tokens = res;
        } catch (err) {
            $createSnackbar(`Could not load authentication tokens: ${err}`);
        }
        loading = false;
    }
    onMount(fetchUserTokens);
</script>

{#each tokens as token (token.token)}
    <DeleteToken
        token={token.token}
        open={tokensForDeletion.includes(token.token)}
        on:delete={() =>
            (tokens = tokens.filter((t) => t.token !== token.token))}
    />
{/each}

<AddToken
    bind:open={addTokenOpen}
    on:create={(e) =>
        (tokens = [
            ...tokens,
            {
                user: $data.userData.user.username,
                token: e.detail.token,
                data: {
                    label: e.detail.label,
                },
            },
        ])}
/>

<div class="security">
    <Progress bind:loading type="linear" />
    <div class="security__tokens">
        <div class="security__tokens__header">
            <div class="securityy__tokens__header__left">
                <h6>Authentication Tokens</h6>
                <span class="text-hint"
                    >Allow you to login via Smarthome apps without a password</span
                >
            </div>
            {#if tokens.length > 0}
                <Button
                    on:click={() => (addTokenOpen = true)}>Add Token</Button
                >
            {/if}
        </div>
        <div class="security__tokens__table">
            {#if tokens.length > 0}
                <DataTable class="security__tokens__table__component">
                    <Head>
                        <Row>
                            <Cell>Visible</Cell>
                            <Cell>Label</Cell>
                            <Cell>Token</Cell>
                            <Cell />
                        </Row>
                    </Head>
                    <Body>
                        {#each tokens as token (token.token)}
                            <Row>
                                <Cell>
                                    <IconButton
                                        class="material-icons"
                                        on:click={() => {
                                            if (
                                                visibleTokens.includes(
                                                    token.token
                                                )
                                            )
                                                visibleTokens =
                                                    visibleTokens.filter(
                                                        (t) => t !== token.token
                                                    );
                                            else
                                                visibleTokens = [
                                                    ...visibleTokens,
                                                    token.token,
                                                ];
                                        }}
                                    >
                                        {visibleTokens.includes(token.token)
                                            ? "visibility_off"
                                            : "visibility"}
                                    </IconButton>
                                </Cell>
                                <Cell>{token.data.label}</Cell>
                                <Cell>
                                    <code
                                        class="security__tokens__table__token-code"
                                    >
                                        {visibleTokens.includes(token.token)
                                            ? token.token
                                            : "*".repeat(token.token.length)}
                                    </code>
                                </Cell>
                                <Cell>
                                    <IconButton
                                        class="material-icons"
                                        on:click={() =>
                                            (tokensForDeletion = [
                                                ...tokensForDeletion,
                                                token.token,
                                            ])}>delete</IconButton
                                    >
                                </Cell>
                            </Row>
                        {/each}
                    </Body>
                </DataTable>
            {:else}
                <div class="security__tokens__table__empty">
                    <i class="material-icons">key_off</i>
                    <h6 class="text-hint">No Tokens</h6>
                    <Button
                        variant="raised"
                        on:click={() => (addTokenOpen = true)}>Add Token</Button
                    >
                </div>
            {/if}
        </div>
    </div>
</div>

<style lang="scss">
    .security {
        padding: 1rem 1.5rem;

        &__tokens {
            &__header {
                display: flex;
                align-items: flex-end;
                justify-content: space-between;

                h6 {
                    margin: 0;
                    margin-top: 0.25rem;
                }
                span {
                    font-size: 0.9rem;
                }
            }

            &__table {
                margin-top: 1rem;

                &__empty {
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    gap: 1rem;
                    i {
                        font-size: 5rem;
                        color: var(--clr-text-disabled);
                    }
                }

                :global &__component {
                    width: 100%;
                    background-color: var(--clr-height-0-3);
                    height: 20rem;
                }

                &__token-code {
                    font-size: 0.7rem;
                    font-family: "Jetbrains Mono", monospace;
                }
            }
        }
    }
</style>
