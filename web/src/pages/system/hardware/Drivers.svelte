<script lang="ts">
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import { Icon } from "@smui/button";
    import Fab from "@smui/fab";
    import CreateDriver from "./CreateDriver.svelte";
    import IconButton from "@smui/icon-button";
    import type { FetchedDriver, CreateDriverReq } from "../driver"
    import { fetchDrivers, createDriver } from "../driver"
    import DriverComponent from "./DriverComponent.svelte";

    // Specifies whether the loading indicator should be shown or hidden
    let loading = true;

    // Specifies whether the add driver dialog should be open or closed
    let createDriverOpen = false;

    // Contains all hardware nodes
    let driversLoaded = false;
    let drivers: FetchedDriver[] = []
    // $: if (drivers) console.log('updated drivers')

    async function refresh() {
        loading = true
        drivers = await fetchDrivers()
        loading = false
    }

    async function saveDriverConfig(driverVendorId: string, driverModelId: string, configuredData: any) {
        loading = true

        try {
            let res = await fetch(
                '/api/system/hardware/driver/configure', {
                    method: "PUT",
                    body: JSON.stringify({
                        driver: {
                            vendorId: driverVendorId,
                            modelId: driverModelId,
                        },
                        data: configuredData
                    })
                }
            )

            if (res.status !== 200) {
                let msg  = await res.json()
                throw `${msg.message}: ${msg.error}`
            }
        } catch (err) {
            $createSnackbar(`Saving driver configuration failed: ${err}`)
        }

        loading = false
    }

    async function createDriverWrapper(data: CreateDriverReq) {
        loading = true
        await createDriver(data)
        loading = false

        await refresh()
    }

    async function deleteDriver(vendorId: string, modelId: string) {
        try {
            const res = await fetch(
                '/api/system/hardware/driver/delete',
                {
                    method: 'DELETE',
                    body: JSON.stringify({ vendorId, modelId })
                },
            );
            if (res.status !== 200) {
                let msg  = await res.json()
                throw `${msg.message}: ${msg.error}`
            }

            drivers = drivers.filter(d => d.driver.vendorId !== vendorId && d.driver.modelId != modelId)
        } catch (error) {
            $createSnackbar(`Deleting driver failed: ${error}`)
        }
    }

    onMount(refresh);
</script>

<CreateDriver
    bind:open={createDriverOpen}
    on:create={(e) => createDriverWrapper(e.detail)}
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
                    on:click={refresh}
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
        <div class="hardware__drivers">
            {#if drivers.length === 0 && driversLoaded}
                <i class="material-icons text-disabled">dns</i>
                <span class="text-hint">No installed drivers </span>
            {:else}
                {#each drivers as driver}
                    <DriverComponent
                        bind:driver
                        on:save={(e) => saveDriverConfig(driver.driver.vendorId, driver.driver.modelId, e.detail)}
                        on:delete={deleteDriver(driver.driver.vendorId, driver.driver.modelId)}
                    />
                {/each}
                <div class="hardware__drivers__placeholder" />
                <div class="hardware__drivers__placeholder" />
            {/if}
        </div>
    </div>
</div>

<style lang="scss">
    @use './drivers.scss' as *;

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

        // Driver list
        &__drivers {
            display: flex;
            flex-wrap: wrap;
            gap: 1.5rem;

            &__placeholder {
                flex-grow: 1;
                width: $driver-width;
            }
        }
    }
</style>
