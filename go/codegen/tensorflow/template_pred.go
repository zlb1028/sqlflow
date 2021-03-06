// Copyright 2020 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorflow

import "sqlflow.org/sqlflow/go/ir"

type predFiller struct {
	DataSource  string
	Select      string
	ResultTable string
	// below members comes from trainStmt
	Estimator         string
	FieldDescs        map[string][]*ir.FieldDesc
	FeatureColumnCode string
	TrainLabelMeta    *ir.FieldDesc
	PredLabelMeta     *ir.FieldDesc
	ModelParams       map[string]interface{}
	PredParams        map[string]interface{}
	Save              string
	HDFSNameNodeAddr  string
	HiveLocation      string
	HDFSUser          string
	HDFSPass          string
}

const tfPredTemplateText = `
import tensorflow as tf
import runtime
from runtime.tensorflow.predict import pred
from runtime.tensorflow.get_tf_version import tf_is_version2
from tensorflow.estimator import DNNClassifier, DNNRegressor, LinearClassifier, LinearRegressor, BoostedTreesClassifier, BoostedTreesRegressor, DNNLinearCombinedClassifier, DNNLinearCombinedRegressor
try:
    import sqlflow_models
except:
    pass

feature_column_names = [{{range $target, $desclist := .FieldDescs}}{{range $desclist}}
"{{.Name}}",
{{end}}{{end}}]

# feature_column_names_map is used to determine the order of feature columns of each target:
# e.g. when using DNNLinearCombinedClassifier
feature_column_names_map = dict()
{{range $target, $desclist := .FieldDescs}}
feature_column_names_map["{{$target}}"] = [{{range $desclist}}"{{.Name}}",{{end}}]
{{end}}
    

feature_metas = dict()
{{ range $target, $desclist := .FieldDescs }}
{{ range $value := $desclist }}
feature_metas["{{$value.Name}}"] = {
    "feature_name": "{{$value.Name}}",
    "dtype": "{{$value.DType | DTypeToString}}",
    "delimiter": "{{$value.Delimiter}}",
    "format": "{{$value.Format}}",
    "shape": {{$value.Shape | intArrayToJSONString}},
    "is_sparse": "{{$value.IsSparse}}" == "true",
    "dtype_weight": "{{$value.DTypeWeight | DTypeToString}}",
    "delimiter_kv": "{{$value.DelimiterKV}}"
}
{{end}}
{{end}}

label_meta = {
    "feature_name": "{{.PredLabelMeta.Name}}",
    "dtype": "{{.PredLabelMeta.DType | DTypeToString}}",
    "delimiter": "{{.PredLabelMeta.Delimiter}}",
    "shape": {{.PredLabelMeta.Shape | intArrayToJSONString}},
    "is_sparse": "{{.PredLabelMeta.IsSparse}}" == "true"
}

train_label_meta = {
    "feature_name": "{{.TrainLabelMeta.Name}}",
    "dtype": "{{.TrainLabelMeta.DType | DTypeToString}}",
    "delimiter": "{{.TrainLabelMeta.Delimiter}}",
    "shape": {{.TrainLabelMeta.Shape | intArrayToJSONString}},
    "is_sparse": "{{.TrainLabelMeta.IsSparse}}" == "true"
}

model_params=dict()
{{range $k, $v := .ModelParams}}
model_params["{{$k}}"]={{$v | attrToPythonValue}}
{{end}}

pred_params=dict()
{{range $k, $v := .PredParams}}
pred_params["{{$k}}"]={{$v | attrToPythonValue}}
{{end}}

feature_columns = {{.FeatureColumnCode}}

pred(datasource="{{.DataSource}}",
     estimator_string="""{{.Estimator}}""",
     select="""{{.Select}}""",
     result_table="{{.ResultTable}}",
     feature_columns=feature_columns,
     feature_column_names=feature_column_names,
     feature_column_names_map=feature_column_names_map,
     train_label_name=train_label_meta["feature_name"],
     result_col_name=label_meta["feature_name"],
     feature_metas=feature_metas,
     model_params=model_params,
     pred_params=pred_params,
     save="{{.Save}}",
     batch_size=1)
`
