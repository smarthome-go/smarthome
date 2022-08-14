<script lang="ts">
    import Box from "./Box.svelte";
    import { onMount } from "svelte";
    import { createSnackbar, data as userData } from "../../global";
    import "chartjs-adapter-date-fns";

    import {
        Chart,
        Title,
        Tooltip,
        Legend,
        LineElement,
        LinearScale,
        PointElement,
        CategoryScale,
        LineController,
        TimeScale,
        BarController,
        BarElement,
        Filler,
        ChartConfiguration,
    } from "chart.js";

    // Register all chart modules
    Chart.register(
        Title,
        Tooltip,
        Filler,
        Legend,
        LineElement,
        LineController,
        BarController,
        BarElement,
        LinearScale,
        PointElement,
        CategoryScale,
        TimeScale
    );

    // Whether the progress indicator should be active
    export let loading = false;

    // Holds the chart's HTML canvas
    let chartCanvas: HTMLCanvasElement = undefined;

    // Saves the power usage time points
    let powerUsageData = [];

    // Fetches the power usage data and places it in `powerUsageData`
    async function fetchData() {
        try {
            const res = await (await fetch("/api/power/usage/day")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Transform the data into valid chart.js data
            for (let record of res) {
                // Because of weird timezones, two hours must be added
                let d = new Date(record.time - 1000 * 60 * 60 * 2);
                // Add the current timestamp to the data
                powerUsageData.push({
                    x: d,
                    y: record.on.watts,
                });
            }
            // Always add a blank point at the end (useful in case there have been no power changes recently)
            if (res.length > 0) {
                let d = new Date();
                // Add the current time at the end
                powerUsageData.push({
                    x: d,
                    y: res[res.length - 1].on.watts,
                });
            }
        } catch (err) {
            $createSnackbar(`Failed to load power usage history: ${err}`);
        }
    }

    function hexToRgb(hex: string): [number, number, number] {
        var bigint = parseInt(hex, 16);
        var r = (bigint >> 16) & 255;
        var g = (bigint >> 8) & 255;
        var b = bigint & 255;

        return [r, g, b];
    }

    const options: ChartConfiguration = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            // @ts-ignore
            legend: {
                display: false,
            },
            tooltip: {
                callbacks: {
                    label: (item: any) =>
                        `${item.parsed.y} Watt${
                            item.parsed.y === 1 ? "" : "s"
                        }`,
                },
            },
        },
        scales: {
            x: {
                type: "time",
                time: {
                    unit: "hour",
                },
            },
        },
    };

    onMount(async () => {
        // Fetch the power usage data
        await fetchData();

        // Create a new canvas context
        let ctx = chartCanvas.getContext("2d");

        // Get  RGB values from user's primary color
        const rgb = hexToRgb(
            $userData.userData.user.darkTheme
                ? $userData.userData.user.primaryColorDark.substring(1)
                : $userData.userData.user.primaryColorLight.substring(1)
        );

        // Specify the gradient which is used to fill the area below the graph
        let gradient = ctx.createLinearGradient(0, 0, 0, 400);
        gradient.addColorStop(
            0.125,
            `rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, 0.3)`
        );
        gradient.addColorStop(0.65, "rgb(0, 0, 100, 0.125)");
        //gradient.addColorStop(0.7, "rgb(0,  0, 100, 0.03)");

        // Dataset configuration
        let data = {
            datasets: [
                {
                    // The area below the graph should be filled
                    fill: true,
                    // Use the gradient as the background color
                    backgroundColor: gradient,
                    // TODO: tweak this value to make it look the best
                    lineTension: 0.0,
                    // Make the graph's line appear in the user's primary color
                    borderColor: $userData.userData.user.darkTheme
                        ? $userData.userData.user.primaryColorDark
                        : $userData.userData.user.primaryColorLight,
                    // Make the point border color appear in the user's primary color
                    pointBorderColor: $userData.userData.user.darkTheme
                        ? $userData.userData.user.primaryColorDark
                        : $userData.userData.user.primaryColorLight,
                    // The inner color of each point
                    pointBackgroundColor: "#252525",
                    // How wide the point's border should be
                    pointBorderWidth: 1,
                    // Sets the point's default width
                    pointRadius: 3.14159265,
                    // Threshold on which `hover` is triggered
                    pointHitRadius: 15,
                    // Widen the point on hover
                    pointHoverRadius: 6,
                    // Darken the point's inner color on hover
                    pointHoverBackgroundColor: "#222",
                    // Increment the point's border width on hover
                    pointHoverBorderWidth: 2,
                    // Could be used to make the graph's line dashed
                    //borderDash: [4],
                    data: powerUsageData,
                },
            ],
        };

        // Create the chart
        let myChart = new Chart(ctx, {
            type: "line",
            data,
            // @ts-ignore
            options,
        });

        // Darkmode-specific configuration
        const x = myChart.config.options.scales.x;
        const y = myChart.config.options.scales.y;

        if ($userData.userData.user.darkTheme) {
            // Decrease the X and Y axis' grid opacity
            x.grid.color = "rgba(255, 255, 255, 0.025)";
            y.grid.color = "rgba(100, 100, 100, 0.25)";
            // Darken the tooltip background color
            Chart.defaults.plugins.tooltip.backgroundColor =
                "rgba(50, 50, 50, 0.75)";
            // Remove borders around the X and Y axis
            x.grid.borderColor = "transparent";
            y.grid.borderColor = "transparent";
        } else {
            // Darken the tooltip color
            Chart.defaults.plugins.tooltip.backgroundColor =
                "rgba(0, 0, 0, 0.7)";
            x.grid.borderColor = "#666";
            y.grid.borderColor = "#666";
            x.grid.color = "rgba(0, 0, 0, 0.1)";
            y.grid.color = "rgba(0, 0, 0, 0.1)";
            Chart.defaults.color = "#666";
        }
    });
</script>

<Box bind:loading>
    <span slot="header">Power Usage</span>
    <div class="content" slot="content">
        <canvas
            bind:this={chartCanvas}
            class="chart"
            class:darkMode={$userData.userData.user.darkTheme}
        />
    </div>
</Box>

<style lang="scss">
    .content {
        // The height is fixed to this value to make it appear nice on most screens
        height: 15rem;
    }
    // Chart styling
    .chart {
        background-color: transparent;

        &.darkMode {
            background-color: transparent;
        }
    }
</style>
