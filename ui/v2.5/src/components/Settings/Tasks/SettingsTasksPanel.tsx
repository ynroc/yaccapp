import React, { useState } from "react";
import { useIntl } from "react-intl";
import { LoadingIndicator } from "src/components/Shared";
import { LibraryTasks } from "./LibraryTasks";
import { DataManagementTasks } from "./DataManagementTasks";
import { PluginTasks } from "./PluginTasks";
import { JobTable } from "./JobTable";

export const SettingsTasksPanel: React.FC = () => {
  const intl = useIntl();
  const [isBackupRunning, setIsBackupRunning] = useState<boolean>(false);

  if (isBackupRunning) {
    return (
      <LoadingIndicator
        message={intl.formatMessage({ id: "config.tasks.backing_up_database" })}
      />
    );
  }

  return (
    <div id="tasks-panel">
      <div className="tasks-panel-queue">
        <h1>{intl.formatMessage({ id: "config.tasks.job_queue" })}</h1>
        <JobTable />
      </div>

      <div className="tasks-panel-tasks">
        <LibraryTasks />
        <hr />
        <DataManagementTasks setIsBackupRunning={setIsBackupRunning} />
        <hr />
        <PluginTasks />
      </div>
    </div>
  );
};
