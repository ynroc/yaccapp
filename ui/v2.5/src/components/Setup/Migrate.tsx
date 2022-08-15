import React, { useEffect, useState } from "react";
import { Button, Card, Container, Form } from "react-bootstrap";
import { useIntl, FormattedMessage } from "react-intl";
import * as GQL from "src/core/generated-graphql";
import { useSystemStatus, mutateMigrate } from "src/core/StashService";
import { LoadingIndicator } from "../Shared";

export const Migrate: React.FC = () => {
  const { data: systemStatus, loading } = useSystemStatus();
  const [backupPath, setBackupPath] = useState<string | undefined>();
  const [migrateLoading, setMigrateLoading] = useState(false);
  const [migrateError, setMigrateError] = useState("");

  const intl = useIntl();

  // make suffix based on current time
  const now = new Date()
    .toISOString()
    .replace(/T/g, "_")
    .replace(/-/g, "")
    .replace(/:/g, "")
    .replace(/\..*/, "");
  const defaultBackupPath = systemStatus
    ? `${systemStatus.systemStatus.databasePath}.${systemStatus.systemStatus.databaseSchema}.${now}`
    : "";

  const discordLink = (
    <a href="https://discord.gg/2TsNFKt" target="_blank" rel="noreferrer">
      Discord
    </a>
  );
  const githubLink = (
    <a
      href="https://github.com/stashapp/stash/issues"
      target="_blank"
      rel="noreferrer"
    >
      <FormattedMessage id="setup.github_repository" />
    </a>
  );

  useEffect(() => {
    if (backupPath === undefined && defaultBackupPath) {
      setBackupPath(defaultBackupPath);
    }
  }, [defaultBackupPath, backupPath]);

  // only display setup wizard if system is not setup
  if (loading || !systemStatus) {
    return <LoadingIndicator />;
  }

  if (migrateLoading) {
    return (
      <LoadingIndicator
        message={intl.formatMessage({ id: "setup.migrate.migrating_database" })}
      />
    );
  }

  if (
    systemStatus.systemStatus.status !== GQL.SystemStatusEnum.NeedsMigration
  ) {
    // redirect to main page
    const newURL = new URL("/", window.location.toString());
    window.location.href = newURL.toString();
    return <LoadingIndicator />;
  }

  const status = systemStatus.systemStatus;

  async function onMigrate() {
    try {
      setMigrateLoading(true);
      setMigrateError("");
      await mutateMigrate({
        backupPath: backupPath ?? "",
      });

      const newURL = new URL("/", window.location.toString());
      window.location.href = newURL.toString();
    } catch (e) {
      if (e instanceof Error) setMigrateError(e.message ?? e.toString());
      setMigrateLoading(false);
    }
  }

  function maybeRenderError() {
    if (!migrateError) {
      return;
    }

    return (
      <section>
        <h2 className="text-danger">
          <FormattedMessage id="setup.migrate.migration_failed" />
        </h2>

        <p>
          <FormattedMessage id="setup.migrate.migration_failed_error" />
        </p>

        <Card>
          <pre>{migrateError}</pre>
        </Card>

        <p>
          <FormattedMessage
            id="setup.migrate.migration_failed_help"
            values={{ discordLink, githubLink }}
          />
        </p>
      </section>
    );
  }

  return (
    <Container>
      <h1 className="text-center mb-3">
        <FormattedMessage id="setup.migrate.migration_required" />
      </h1>
      <Card>
        <section>
          <p>
            <FormattedMessage
              id="setup.migrate.schema_too_old"
              values={{
                databaseSchema: <strong>{status.databaseSchema}</strong>,
                appSchema: <strong>{status.appSchema}</strong>,
                strong: (chunks: string) => <strong>{chunks}</strong>,
                code: (chunks: string) => <code>{chunks}</code>,
              }}
            />
          </p>

          <p className="lead text-center my-5">
            <FormattedMessage id="setup.migrate.migration_irreversible_warning" />
          </p>

          <p>
            <FormattedMessage
              id="setup.migrate.backup_recommended"
              values={{
                defaultBackupPath,
                code: (chunks: string) => <code>{chunks}</code>,
              }}
            />
          </p>
        </section>

        <section>
          <Form.Group id="migrate">
            <Form.Label>
              <FormattedMessage id="setup.migrate.backup_database_path_leave_empty_to_disable_backup" />
            </Form.Label>
            <Form.Control
              className="text-input"
              name="backupPath"
              defaultValue={backupPath}
              placeholder={intl.formatMessage({
                id: "setup.paths.database_filename_empty_for_default",
              })}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setBackupPath(e.currentTarget.value)
              }
            />
          </Form.Group>
        </section>

        <section>
          <div className="d-flex justify-content-center">
            <Button variant="primary mx-2 p-5" onClick={() => onMigrate()}>
              <FormattedMessage id="setup.migrate.perform_schema_migration" />
            </Button>
          </div>
        </section>

        {maybeRenderError()}
      </Card>
    </Container>
  );
};

export default Migrate;
