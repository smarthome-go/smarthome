<script lang="ts">
    import Nodes from "./nodes/Nodes.svelte";
    import type { hardwareNode } from "./types";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import { Icon } from "@smui/button";
    import Fab from "@smui/fab";
    import CreateDriver from "./CreateDriver.svelte";
    import IconButton from "@smui/icon-button";
    import type { homescript } from "src/homescript";

    // Specifies whether the loading indicator should be shown or hidden
    let loading = true;

    // Specifies whether the add driver dialog should be open or closed
    let createDriverOpen = false;

    // Contains all hardware nodes
    let driversLoaded = false;
    let drivers: DeviceDriver[] = []

    interface DeviceDriver {
        driver: DriverData,
        homescript: homescript
    }

    interface DriverData {
        vendorId: string,
        modelId: string,
        name: string,
        version: string,
    }

    interface CreateDriver {
        data: DriverData,
        code: string,
    }

    async function fetchDrivers() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/hardware/drivers/list")
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            drivers = res;
            driversLoaded = true;
        } catch (err) {
            $createSnackbar(`Failed to load hardware drivers: ${err}`);
        }
        loading = false;
    }

    // Creates a new hardware node
    async function createDriver(
        data: CreateDriver
    ) {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/hardware/drivers/add", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(data),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            drivers = [
                ...drivers,
                {
                    driver: data.data,
                    homescript: res as homescript
                },
            ];
        } catch (err) {
            $createSnackbar(`Failed to create hardware driver node: ${err}`);
        }
        loading = false;
    }

    onMount(() => fetchDrivers());
</script>

<CreateDriver
    bind:open={createDriverOpen}
    on:create={(e) => /*TODO: create driver*/ console.log(e)}
/>

<div class="hardware">
    <Progress bind:loading />
    <h6>Drivers</h6>

    <div class="hardware__type">
        <div class="hardware__type__label">
            <a
                class="hardware__type__label__name"
                href="https://github.com/smarthome-go/node"
                rel="noopener noreferrer nofollow"
                target="_blank"
                >Drivers (TODO: link to wiki)
            </a>
            <i class="hardware__type__label__icon material-icons">memory</i>
            <div class="hardware__type__label__right">
                <IconButton
                    disabled={loading}
                    class="material-icons"
                    on:click={() => fetchDrivers()}
                    title="Refresh">refresh</IconButton
                >
                <Fab
                    color="primary"
                    mini
                    title="Add Node"
                    on:click={() => (createDriverOpen = true)}
                >
                    <Icon class="material-icons">add</Icon>
                </Fab>
            </div>
        </div>

        <!-->vendor starts here</-->
        <div class="hardware__nodes">
            {#if drivers.length === 0 && driversLoaded}
                <i class="material-icons text-disabled">dns</i>
                <span class="text-hint"> No installed drivers </span>
            {:else}
                {#each drivers as driver}
                    {driver}
                {/each}
            {/if}
        </div>
    </div>
</div>

<style lang="scss">
    // Main list which contains different kinds of manufacturers
    .hardware {
        padding: 1rem 1.5rem;

        h6 {
            margin: 0;
            font-size: 1.1rem;
            color: var(--clr-text-hint);
        }

        &__type {
            &__label {
                display: flex;
                align-items: center;
                gap: 0.4rem;
                margin-top: 1rem;
                margin-bottom: 0.5rem;

                // Any HTML element which can be used to label the coming hardware section
                // Often an `a-tag` which links to a reference page
                &__name {
                    color: var(--clr-text-hint);
                }
                // `i-tag` which contains a MD icon
                &__icon {
                    color: var(--clr-text-hint);
                    font-size: 1.5rem;
                }
                &__right {
                    margin-left: auto;
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                }
            }
        }

        /*
           Vendor-specific styles start here
        */

        // The default hardware node device type
        &__nodes {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
        }
    }
</style>
