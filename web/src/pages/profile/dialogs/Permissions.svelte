<script lang="ts">
    import Button, { Label } from "@smui/button";

    import DataTable, { Head, Body, Row, Cell } from "@smui/data-table";
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";
    import IconButton from "@smui/icon-button";
    import LinearProgress from "@smui/linear-progress";
    import { onMount } from "svelte";
    import { data, createSnackbar } from "../../../global";

    export let open = false;

    interface Permission {
        permission: string;
        name: string;
        description: string;
    }

    let permissions: Permission[] = [];
    let permissionsLoaded = false;
    async function fetchAllPermissions() {
        try {
            const res = await (await fetch("/api/permissions/list/all")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            permissions = res.filter((p: Permission) =>
                $data.userData.permissions.includes(p.permission)
            );
            permissionsLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load system permissions: ${err}`);
        }
    }

    onMount(fetchAllPermissions);
</script>

<Dialog
    bind:open
    fullscreen
    aria-labelledby="fullscreen-title"
    aria-describedby="fullscreen-content"
>
    <Header>
        <Title id="fullscreen-title">Your Permissions</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="fullscreen-content">
        <div class="misc">
            <div class="misc__permissions">
                <DataTable
                    table$aria-label="Your Permissions"
                    style="width: 100%;"
                >
                    <Head>
                        <Row>
                            <Cell>ID</Cell>
                            <Cell>Name</Cell>
                            <Cell style="width: 100%;">Description</Cell>
                        </Row>
                    </Head>
                    <Body>
                        {#each permissions as permission (permission.permission)}
                            <Row>
                                <Cell>{permission.permission}</Cell>
                                <Cell>{permission.name}</Cell>
                                <Cell>{permission.description}</Cell>
                            </Row>
                        {/each}
                    </Body>

                    <LinearProgress
                        indeterminate
                        bind:closed={permissionsLoaded}
                        aria-label="Permissions are being loaded..."
                        slot="progress"
                    />
                </DataTable>
            </div>
        </div>
    </Content>
    <Actions>
        <Button defaultAction>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .misc {
        &__permissions {
            width: 100%;
        }
    }
</style>
