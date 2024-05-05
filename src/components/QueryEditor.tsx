import React, { ChangeEvent } from 'react';
import { InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {

  const onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    query.uuid = crypto.randomUUID()
    query.selectQuery = event.target.value
    onChange({...query});
  };
 
  const { selectQuery } = query;

  return (
    <div style={{ width: '100%' }}>
      <InlineField label="Query" labelWidth={16} tooltip="SQL Query">
        <Input 
          style={{ width: '100%' }}
          id="query-editor-query-text"
          onChange={onQueryTextChange}
          value={selectQuery || ''}
          required
          placeholder="Enter a query"
        />
      </InlineField>
    </div>
  );
}
