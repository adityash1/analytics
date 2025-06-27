import { TrendingUp } from "lucide-react";
import { LabelList, Pie, PieChart } from "recharts";
import "./App.css";

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";

import axios from "axios";
import { useEffect, useState } from "react";
import useSWRMutation from "swr/mutation";

const chartConfig = {
  count: {
    label: "Count",
  },
  windows: {
    label: "Windows",
    color: "var(--chart-1)",
  },
  linux: {
    label: "Linux",
    color: "var(--chart-2)",
  },
  macos: {
    label: "macOS",
    color: "var(--chart-3)",
  },
  ios: {
    label: "iOS",
    color: "var(--chart-4)",
  },
  android: {
    label: "Android",
    color: "var(--chart-5)",
  },
  fallback: {
    label: "Other",
    color: "var(--chart-muted)",
  },
} satisfies ChartConfig;

interface ApiDataItem {
  occuredAt: number;
  value: string;
  count: number;
}

interface FormattedChartDataItem {
  os: string;
  count: number;
  fill: string;
}

type AnalyticsPayload = {
  What: number;
  SiteID: string;
  Start: number;
  End: number;
};

const postAnalytics = async (
  url: string,
  { arg }: { arg: AnalyticsPayload }
) => {
  const response = await axios.post<ApiDataItem[]>(url, arg, {
    headers: {
      "X-API-KEY": "dev",
    },
  });
  return response.data;
};

const payload = {
  What: 6,
  SiteID: "news-corp",
  Start: 20250625,
  End: 20250627,
};

function App() {
  const { trigger, data, error, isMutating } = useSWRMutation(
    "http://localhost:9876/stats",
    postAnalytics
  );

  const [formattedChartData, setFormattedChartData] = useState<
    FormattedChartDataItem[]
  >([]);

  useEffect(() => {
    trigger(payload);
  }, [trigger]);

  useEffect(() => {
    if (data && Array.isArray(data)) {
      const formatted = data.map((item: ApiDataItem) => {
        const key = item.value.toLowerCase().replace(/\s+/g, "");
        const configEntry =
          chartConfig[key as keyof Omit<typeof chartConfig, "count">] ||
          chartConfig.fallback;
        return {
          os: item.value,
          count: item.count,
          fill: configEntry.color,
        };
      });
      setFormattedChartData(formatted);
    }
  }, [data]);

  return (
    <>
      <div>
        {isMutating && <p>Loading analytics...</p>}
        {error && <p style={{ color: "red" }}>Error: {error.message}</p>}
      </div>
      <Card className="flex flex-col">
        <CardHeader className="items-center pb-0">
          <CardTitle>Pie Chart - OS Distribution</CardTitle>
          <CardDescription>Based on recent activity</CardDescription>
        </CardHeader>
        <CardContent className="flex-1 pb-0">
          <ChartContainer
            config={chartConfig}
            className="mx-auto aspect-square max-h-[250px] [&_.recharts-text]:fill-background"
          >
            <PieChart>
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent nameKey="count" hideLabel />}
              />
              <Pie data={formattedChartData} dataKey="count" nameKey="os">
                <LabelList
                  dataKey="os"
                  className="fill-background"
                  stroke="none"
                  fontSize={12}
                  formatter={(value: string) => {
                    const key = value.toLowerCase().replace(/\s+/g, "");
                    return (
                      chartConfig[key as keyof typeof chartConfig]?.label ||
                      value
                    );
                  }}
                />
              </Pie>
            </PieChart>
          </ChartContainer>
        </CardContent>
        <CardFooter className="flex-col gap-2 text-sm">
          <div className="flex items-center gap-2 font-medium leading-none">
            Showing OS distribution <TrendingUp className="h-4 w-4" />
          </div>
          <div className="leading-none text-muted-foreground">
            Updated based on the latest fetch
          </div>
        </CardFooter>
      </Card>
    </>
  );
}

export default App;
