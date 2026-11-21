import QtQuick
import Quickshell
import qs.Commons
import qs.Ui

// blackpie's icon in Omarchy's bar, put there by blackpie (Settings → Bar icon).
// A click drops down its panel under the bar, where the click was; a right
// click opens its window. blackpie is started if it isn't running.
BarWidget {
  id: root
  moduleName: "useblackpie.blackpie"

  // written in by blackpie when it puts the widget here
  readonly property string blackpie: "__BLACKPIE__"

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: String.fromCodePoint(0xf15c6) // nf-md-bird
    tooltipText: "blackpie"
    onPressed: function(b) {
      Quickshell.execDetached(b === Qt.RightButton ? [root.blackpie, "app"] : [root.blackpie, "panel"])
    }
  }
}
