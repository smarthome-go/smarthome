<script lang="ts">
    import Tab, { Icon, Label } from "@smui/tab";
    import TabBar from "@smui/tab-bar";
    import { createSnackbar, sleep } from "../../../global";
    import HmsEditor from "../../../components/Homescript/HmsEditor/HmsEditor.svelte";
    import HmsInputsReset from "./HMSInputsReset.svelte";
    import type { homescript } from "../../../homescript";
    import HmsSelector from "../../../components/Homescript/HmsSelector.svelte";
    import Progress from "../../../components/Progress.svelte";

    /*
        //// Tabs (active mode selection) ////
    */

    // Specifies which mode is currently being used for editing
    let active: "hms" | "switches" | "code" = "switches";

    // Saves the last mode in the HMS
    // Uses a header comment which is evaluated and parsed (see below for reference)
    let activeInCode: "hms" | "switches" | "code" = "hms";

    // Saves the tab data for the editor type selection
    const tabs: string[] = ["hms", "switches", "code"];
    const tabData: { label: "hms" | "switches" | "code"; icon: string }[] = [
        {
            label: "hms",
            icon: "list",
        },
        {
            label: "switches",
            icon: "power",
        },
        {
            label: "code",
            icon: "code",
        },
    ];

    /*
        //// Homescripts ////
        Used for when the active mode is set to `hms`
    */

    // Load the Homescripts if required
    $: if (active === "hms" && !homescriptsLoaded && !homescriptsLoading)
        loadHomescripts();

    // Saves the Homescripts which are available to the current user
    // Used for displaying the HMS selector
    let homescripts: homescript[] = [];
    let homescriptsLoaded: boolean = false;
    let homescriptsLoading: boolean = false;

    // Specifies which Homescript should be executed
    // Homescript code will later be genereated reactively
    let selectedHMS: string = "";

    // Update the selected Homescript inside the code
    $: if (active === "hms" && activeInCode == "hms" && selectedHMS != "")
        setHMSInCode();

    function setHMSInCode() {
        if (selectedHMS != "")
            code = code.split("\n")[0] + `\nexec("${selectedHMS}")`;
    }

    function getHMSFromCode() {
        try {
            selectedHMS = code.split('exec("')[1].split('")')[0];
        } catch (err) {
            if (homescripts.length > 0) {
                selectedHMS = homescripts[0].data.id;
            } else {
                selectedHMS = "";
            }
            setHMSInCode();
        }
    }

    // Fetches the user's Homescripts for the HMS selectot
    async function loadHomescripts() {
        homescriptsLoading = true;
        try {
            let res = await (
                await fetch("/api/homescript/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Assign to the HMS list
            homescripts = res;
            // Signal that the HMS are loaded
            homescriptsLoaded = true;
            // Update the selected HMS from the code
            if (active === "hms" && activeInCode == "hms") getHMSFromCode();
        } catch (err) {
            $createSnackbar(`Could not load Homescripts: ${err}`);
        }
        homescriptsLoading = false;
    }

    /*
        //// Switches ////
        Used for when the active mode is set to `switches`
    */
    // Load the switches if required
    $: if (active === "switches" && !switchesLoaded && !switchesLoading)
        loadSwitches();

    // Saves the switches which are available to the current user
    // Used for displaying the switch selector
    let switches: SwitchResponse[];
    let switchesLoaded: boolean = false;
    let switchesLoading: boolean = false;

    interface SwitchResponse {
        id: string;
        name: string;
        powerOn: boolean;
        watts: number;
    }

    // Loads the user's personal switches
    async function loadSwitches() {
        try {
            const res = await (await fetch("/api/switch/list/personal")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            switches = res;
            switchesLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load switches: ${err}`);
        }
    }

    /*
        //// Code + active modes ////
    */
    // Is bound to the HMS editors
    export let code: string = `#active_mode:${active}`;

    // Updates the active-code mode every time the underlying code changes
    $: if (code !== undefined) {
        activeInCode = getModeFromCode();
        active = activeInCode;
    }

    // Parses the code's first line in order to return the active code
    // If the function fails to do so, it returns the current active mode and displays an error-message to the user (who must have messed up)
    function getModeFromCode(): "hms" | "switches" | "code" {
        switch (code.split("\n")[0].split("#active_mode:")[1]) {
            case "hms":
                return "hms";
            case "switches":
                return "switches";
            case "code":
                return "code";
            default:
                $createSnackbar("The first line must not be edited");
                setModeInCode(active);
                return active;
        }
    }

    // Updates the code's header comment to use a given mode
    function setModeInCode(mode: "hms" | "switches" | "code") {
        let codeTemp = code.split("\n");
        codeTemp[0] = `#active_mode:${mode}`;
        code = codeTemp.join("\n");
    }

    // Reset the code using a given mode
    function resetCode(mode: "hms" | "switches" | "code") {
        code = `#active_mode:${mode}`;
    }
</script>

<div class="main">
    <div class="main__header">
        <TabBar {tabs} let:tab bind:active>
            <Tab {tab}>
                <Icon class="material-icons"
                    >{tabData.find((t) => t.label === tab).icon}</Icon
                >
                <Label>{tab}</Label>
            </Tab>
        </TabBar>
    </div>
    <div class="main__editor" class:hms={active === "hms"}>
        {#if active === "hms"}
            {#if activeInCode !== active}
                <HmsInputsReset
                    {active}
                    {activeInCode}
                    icon="auto_fix_off"
                    on:reset={() => resetCode(active)}
                />
            {:else}
                <HmsSelector {homescripts} bind:selection={selectedHMS} />
            {/if}
        {:else if active === "switches"}
            {#if activeInCode !== active}
                <HmsInputsReset
                    {active}
                    {activeInCode}
                    icon="auto_fix_off"
                    on:reset={() => resetCode(active)}
                />
            {:else if switchesLoaded}
                {#each switches as sw (sw.id)}
                    <h6>{sw.name}</h6>
                {/each}
            {:else}
                <Progress type="circular" loading={true} />
            {/if}
        {:else if activeInCode !== active}
            <HmsInputsReset
                {active}
                {activeInCode}
                icon="code_off"
                on:reset={() => resetCode(active)}
            />
        {:else}
            <HmsEditor registerCtrlSCatcher bind:code />
        {/if}
    </div>
</div>

<style lang="scss">
    .main {
        &__editor {
            height: 20rem;

            &.hms {
                height: 100%;
                min-height: 20rem;
            }

            margin-top: 1rem;
            overflow: auto;
        }
    }
</style>
